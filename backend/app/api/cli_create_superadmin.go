package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/pressly/goose/v3"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

// runCreateSuperAdminCLI is the only supported way to create a SuperAdmin.
// It intentionally does not expose a UI or API endpoint for this account type.
func runCreateSuperAdminCLI(args []string) (handled bool, exitCode int) {
	if len(args) < 2 || args[1] != "create-superadmin" {
		return false, 0
	}

	fs := flag.NewFlagSet("create-superadmin", flag.ContinueOnError)
	name := fs.String("name", "", "Display name")
	email := fs.String("email", "", "Email address")
	password := fs.String("password", "", "Password")
	groupName := fs.String("group-name", "", "Optional collection name")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: homebox create-superadmin --name=<name> --email=<address> --password=<password>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Creates a SuperAdmin account directly in the database. This escape hatch intentionally has no UI.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args[2:]); err != nil {
		return true, 2
	}

	trimmedName := strings.TrimSpace(*name)
	trimmedEmail := strings.ToLower(strings.TrimSpace(*email))
	if trimmedName == "" || trimmedEmail == "" || *password == "" {
		fs.Usage()
		return true, 2
	}

	cfg, err := config.New(build(), "Homebox inventory management system")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return true, 1
	}
	if err := createSuperAdminOffline(cfg, trimmedName, trimmedEmail, *password, strings.TrimSpace(*groupName)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create SuperAdmin: %v\n", err)
		return true, 1
	}

	fmt.Printf("SuperAdmin created: %s\n", trimmedEmail)
	return true, 0
}

func createSuperAdminOffline(cfg *config.Config, name, email, password, groupName string) error {
	databaseURL, err := setupDatabaseURL(cfg)
	if err != nil {
		return fmt.Errorf("setup database url: %w", err)
	}

	driver := strings.ToLower(cfg.Database.Driver)
	var driverName, dialectName string
	switch driver {
	case config.DriverPostgres:
		driverName = "pgx"
		dialectName = dialect.Postgres
	case config.DriverSqlite3, "sqlite":
		driverName = "sqlite3"
		dialectName = dialect.SQLite
	default:
		return fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialectName, db)))
	defer func() { _ = client.Close() }()

	migrationsFs, err := migrations.Migrations(driver)
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}
	goose.SetBaseFS(migrationsFs)
	if err := goose.SetDialect(driver); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(client.Sql(), driver); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	ctx := context.Background()
	if exists, err := client.User.Query().Where(user.EmailEqualFold(email)).Exist(ctx); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("an account with email %s already exists", email)
	}

	passwordHash, err := hasher.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	collection, err := tx.Group.Query().First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return err
		}
		nameForGroup := strings.TrimSpace(groupName)
		if nameForGroup == "" {
			nameForGroup = "SuperAdmin Collection"
		}
		collection, err = tx.Group.Create().SetName(nameForGroup).SetCurrency("usd").Save(ctx)
		if err != nil {
			return err
		}
	}

	superAdmin, err := tx.User.Create().
		SetName(name).
		SetEmail(email).
		SetPassword(passwordHash).
		SetIsSuperuser(true).
		SetSuperuser(true).
		SetAccountRole(user.AccountRoleSuperadmin).
		SetDefaultGroupID(collection.ID).
		Save(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.UserGroup.Create().
		SetUserID(superAdmin.ID).
		SetGroupID(collection.ID).
		SetRole(usergroup.RoleOwner).
		Save(ctx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}
