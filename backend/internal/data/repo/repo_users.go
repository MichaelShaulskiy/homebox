package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository struct {
	db *ent.Client
}

// normalizeEmail canonicalizes an email address for storage and lookup. Emails are
// treated case-insensitively throughout the app (login uses a case-folding lookup),
// but the database UNIQUE constraint on users.email is case-sensitive. Storing a
// lowercased, trimmed form makes that constraint reject case-variant duplicates —
// without it an attacker could register USER@EXAMPLE.COM against an existing
// user@example.com, after which the case-insensitive login lookup matches >1 row
// and denies both accounts access.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

type (
	// UserCreate is the Data object contain the requirements of creating a user
	// in the database. It should to create users from an API unless the user has
	// rights to create SuperUsers. For regular user in data use the UserIn struct.
	UserCreate struct {
		Name           string           `json:"name"`
		Email          string           `json:"email"`
		Password       *string          `json:"password"`
		IsSuperuser    bool             `json:"isSuperUser"`
		AccountRole    user.AccountRole `json:"accountRole"`
		DefaultGroupID uuid.UUID        `json:"defaultGroupID"`
		// IsOwner controls the role on the membership row created for
		// (user, DefaultGroupID). It does not grant any cross-group privilege.
		IsOwner bool `json:"isOwner"`
	}

	UserUpdate struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	UserAccountCreate struct {
		Name          string           `json:"name"`
		Email         string           `json:"email"`
		Password      string           `json:"password"`
		AccountRole   user.AccountRole `json:"accountRole"`
		SupervisorIDs []uuid.UUID      `json:"supervisorIds"`
	}

	UserAccountRoleUpdate struct {
		AccountRole user.AccountRole `json:"accountRole"`
	}

	UserAccountSupervisorsUpdate struct {
		SupervisorIDs []uuid.UUID `json:"supervisorIds"`
	}

	UserOut struct {
		ID             uuid.UUID        `json:"id"`
		Name           string           `json:"name"`
		Email          string           `json:"email"`
		IsSuperuser    bool             `json:"isSuperuser"`
		AccountRole    user.AccountRole `json:"accountRole"`
		SupervisorIDs  []uuid.UUID      `json:"supervisorIds"`
		DefaultGroupID uuid.UUID        `json:"defaultGroupId"`
		GroupIDs       []uuid.UUID      `json:"groupIds"`
		PasswordHash   string           `json:"-"`
		OidcIssuer     *string          `json:"oidcIssuer"`
		OidcSubject    *string          `json:"oidcSubject"`
	}

	UserSummary struct {
		Name          string           `json:"name"`
		Email         string           `json:"email"`
		ID            uuid.UUID        `json:"id"`
		AccountRole   user.AccountRole `json:"accountRole"`
		SupervisorIDs []uuid.UUID      `json:"supervisorIds"`
	}
)

var (
	mapUserOutErr      = mapTErrFunc(mapUserOut)
	mapUsersOutErr     = mapTEachErrFunc(mapUserOut)
	mapUsersSummaryErr = mapTEachErrFunc(mapUserSummary)
)

func mapUserOut(user *ent.User) UserOut {
	return UserOut{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		IsSuperuser: user.IsSuperuser,
		AccountRole: user.AccountRole,
		SupervisorIDs: lo.Map(user.Edges.Supervisors, func(supervisor *ent.User, _ int) uuid.UUID {
			return supervisor.ID
		}),
		DefaultGroupID: lo.FromPtrOr(user.DefaultGroupID, uuid.Nil),
		GroupIDs: lo.Map(user.Edges.Groups, func(g *ent.Group, _ int) uuid.UUID {
			return g.ID
		}),
		PasswordHash: lo.FromPtrOr(user.Password, ""),
		OidcIssuer:   user.OidcIssuer,
		OidcSubject:  user.OidcSubject,
	}
}

func mapUserSummary(user *ent.User) UserSummary {
	return UserSummary{
		Name:        user.Name,
		Email:       user.Email,
		ID:          user.ID,
		AccountRole: user.AccountRole,
		SupervisorIDs: lo.Map(user.Edges.Supervisors, func(supervisor *ent.User, _ int) uuid.UUID {
			return supervisor.ID
		}),
	}
}

func userSpanAttrs(out UserOut) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("user.id", out.ID.String()),
		attribute.String("user.default_group_id", out.DefaultGroupID.String()),
		attribute.Int("user.groups.count", len(out.GroupIDs)),
		attribute.Bool("user.is_superuser", out.IsSuperuser),
		attribute.Bool("user.has_password_hash", out.PasswordHash != ""),
		attribute.Bool("user.has_oidc", out.OidcIssuer != nil && out.OidcSubject != nil),
	}
}

func (r *UserRepository) GetOneID(ctx context.Context, id uuid.UUID) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetOneID",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	out, err := mapUserOutErr(r.db.User.Query().
		Where(user.ID(id)).
		WithGroups().
		WithSupervisors().
		Only(ctx))
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(userSpanAttrs(out)...)
	return out, nil
}

func (r *UserRepository) GetOneEmail(ctx context.Context, email string) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetOneEmail",
		trace.WithAttributes(attribute.Int("user.email.length", len(email))))
	defer span.End()

	out, err := mapUserOutErr(r.db.User.Query().
		Where(user.EmailEqualFold(normalizeEmail(email))).
		WithGroups().
		WithSupervisors().
		Only(ctx),
	)
	if err != nil {
		// "not found" is expected on bad logins; record on the span but don't mark
		// it as an error status so dashboards aren't flooded with red.
		span.SetAttributes(
			attribute.Bool("user.found", false),
			attribute.String("user.lookup.error", err.Error()),
			attribute.Bool("user.lookup.not_found", ent.IsNotFound(err)),
		)
		if !ent.IsNotFound(err) {
			recordSpanError(span, err)
		}
		return out, err
	}
	span.SetAttributes(attribute.Bool("user.found", true))
	span.SetAttributes(userSpanAttrs(out)...)
	return out, nil
}

func (r *UserRepository) GetOneEmailNoEdges(ctx context.Context, email string) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetOneEmailNoEdges",
		trace.WithAttributes(attribute.Int("user.email.length", len(email))))
	defer span.End()

	out, err := mapUserOutErr(r.db.User.Query().
		Where(user.EmailEqualFold(normalizeEmail(email))).
		Only(ctx),
	)
	if err != nil {
		span.SetAttributes(
			attribute.Bool("user.found", false),
			attribute.Bool("user.lookup.not_found", ent.IsNotFound(err)),
		)
		if !ent.IsNotFound(err) {
			recordSpanError(span, err)
		}
		return out, err
	}
	span.SetAttributes(attribute.Bool("user.found", true))
	span.SetAttributes(userSpanAttrs(out)...)
	return out, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetAll")
	defer span.End()

	out, err := mapUsersOutErr(r.db.User.Query().WithGroups().WithSupervisors().All(ctx))
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.Int("users.count", len(out)))
	return out, nil
}

// membershipRole returns the per-membership role to assign for a UserCreate.
func membershipRole(isOwner bool) usergroup.Role {
	if isOwner {
		return usergroup.RoleOwner
	}
	return usergroup.RoleUser
}

// createUserWithMembership inserts the user row and the (user, default_group)
// membership in a single transaction so the user always has exactly one
// membership row when Create returns.
func (r *UserRepository) createUserWithMembership(
	ctx context.Context,
	usr UserCreate,
	configure func(*ent.UserCreate) *ent.UserCreate,
) (uuid.UUID, error) {
	if usr.AccountRole == "" {
		usr.AccountRole = user.AccountRoleUser
	}
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	q := tx.User.
		Create().
		SetName(usr.Name).
		SetEmail(normalizeEmail(usr.Email)).
		SetIsSuperuser(usr.IsSuperuser).
		SetAccountRole(usr.AccountRole).
		SetDefaultGroupID(usr.DefaultGroupID)

	if usr.Password != nil {
		q = q.SetPassword(*usr.Password)
	}
	if configure != nil {
		q = configure(q)
	}

	entUser, err := q.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return uuid.Nil, err
	}

	if _, err := tx.UserGroup.Create().
		SetUserID(entUser.ID).
		SetGroupID(usr.DefaultGroupID).
		SetRole(membershipRole(usr.IsOwner)).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return uuid.Nil, err
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return entUser.ID, nil
}

func (r *UserRepository) Create(ctx context.Context, usr UserCreate) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.Create",
		trace.WithAttributes(
			attribute.String("user.default_group_id", usr.DefaultGroupID.String()),
			attribute.Bool("user.is_superuser", usr.IsSuperuser),
			attribute.Bool("user.is_owner", usr.IsOwner),
			attribute.Bool("user.has_password", usr.Password != nil),
		))
	defer span.End()

	id, err := r.createUserWithMembership(ctx, usr, nil)
	if err != nil {
		recordSpanError(span, err)
		return UserOut{}, err
	}
	span.SetAttributes(attribute.String("user.id", id.String()))

	out, err := r.GetOneID(ctx, id)
	if err != nil {
		recordSpanError(span, err)
	}
	return out, err
}

func (r *UserRepository) CreateWithOIDC(ctx context.Context, usr UserCreate, issuer, subject string) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.CreateWithOIDC",
		trace.WithAttributes(
			attribute.String("user.default_group_id", usr.DefaultGroupID.String()),
			attribute.Bool("user.is_superuser", usr.IsSuperuser),
			attribute.Bool("user.is_owner", usr.IsOwner),
			attribute.Bool("user.has_password", usr.Password != nil),
			attribute.String("oidc.issuer", issuer),
			attribute.Int("oidc.subject.length", len(subject)),
		))
	defer span.End()

	id, err := r.createUserWithMembership(ctx, usr, func(uc *ent.UserCreate) *ent.UserCreate {
		return uc.SetOidcIssuer(issuer).SetOidcSubject(subject)
	})
	if err != nil {
		recordSpanError(span, err)
		return UserOut{}, err
	}
	span.SetAttributes(attribute.String("user.id", id.String()))

	out, err := r.GetOneID(ctx, id)
	if err != nil {
		recordSpanError(span, err)
	}
	return out, err
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, data UserUpdate) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.Update",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	q := r.db.User.Update().
		Where(user.ID(id)).
		SetName(data.Name).
		SetEmail(normalizeEmail(data.Email))

	_, err := q.Save(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) SetAccountRole(ctx context.Context, id uuid.UUID, role user.AccountRole) error {
	if role != user.AccountRoleAdmin && role != user.AccountRoleUser {
		return fmt.Errorf("only admin and user roles may be assigned through the UI")
	}
	target, err := r.db.User.Get(ctx, id)
	if err != nil {
		return err
	}
	if target.AccountRole == user.AccountRoleSuperadmin || target.IsSuperuser || target.Superuser {
		return fmt.Errorf("SuperAdmin accounts may only be managed outside the UI")
	}
	_, err = r.db.User.UpdateOneID(id).
		SetAccountRole(role).
		SetIsSuperuser(false).
		SetSuperuser(false).
		Save(ctx)
	return err
}

func (r *UserRepository) SetSupervisors(ctx context.Context, id uuid.UUID, supervisorIDs []uuid.UUID) error {
	if len(supervisorIDs) > 0 {
		supervisors, err := r.db.User.Query().
			Where(user.IDIn(supervisorIDs...)).
			All(ctx)
		if err != nil {
			return err
		}
		if len(supervisors) != len(supervisorIDs) {
			return fmt.Errorf("one or more supervisors do not exist")
		}
		for _, supervisor := range supervisors {
			if supervisor.AccountRole != user.AccountRoleAdmin &&
				supervisor.AccountRole != user.AccountRoleSuperadmin &&
				!supervisor.IsSuperuser {
				return fmt.Errorf("all supervisors must be admin or superadmin")
			}
		}
	}
	builder := r.db.User.UpdateOneID(id).ClearSupervisors()
	if len(supervisorIDs) > 0 {
		builder.AddSupervisorIDs(supervisorIDs...)
	}
	_, err := builder.Save(ctx)
	return err
}

func (r *UserRepository) GetAccountSupervisors(ctx context.Context, gid uuid.UUID) ([]UserSummary, error) {
	rows, err := r.db.User.Query().
		Where(
			user.HasGroupsWith(group.ID(gid)),
			user.Or(
				user.AccountRoleEQ(user.AccountRoleAdmin),
				user.AccountRoleEQ(user.AccountRoleSuperadmin),
				user.IsSuperuser(true),
			),
		).
		WithSupervisors().
		Order(user.ByName()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserSummary, len(rows))
	for i, row := range rows {
		out[i] = mapUserSummary(row)
	}
	return out, nil
}

func (r *UserRepository) UpdateDefaultGroup(ctx context.Context, id uuid.UUID, groupID uuid.UUID) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.UpdateDefaultGroup",
		trace.WithAttributes(
			attribute.String("user.id", id.String()),
			attribute.String("group.id", groupID.String()),
		))
	defer span.End()

	err := r.db.User.UpdateOneID(id).SetDefaultGroupID(groupID).Exec(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.Delete",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	_, err := r.db.User.Delete().Where(user.ID(id)).Exec(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) DeleteAll(ctx context.Context) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.DeleteAll")
	defer span.End()

	_, err := r.db.User.Delete().Exec(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) GetSuperusers(ctx context.Context) ([]*ent.User, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetSuperusers")
	defer span.End()

	users, err := r.db.User.Query().Where(user.IsSuperuser(true)).All(ctx)
	if err != nil {
		recordSpanError(span, err)
		return nil, err
	}
	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

func (r *UserRepository) ChangePassword(ctx context.Context, uid uuid.UUID, pw string) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.ChangePassword",
		trace.WithAttributes(
			attribute.String("user.id", uid.String()),
			attribute.Int("password.hash.length", len(pw)),
		))
	defer span.End()

	err := r.db.User.UpdateOneID(uid).SetPassword(pw).Exec(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) SetSettings(ctx context.Context, uid uuid.UUID, settings map[string]interface{}) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.SetSettings",
		trace.WithAttributes(
			attribute.String("user.id", uid.String()),
			attribute.Int("settings.keys.count", len(settings)),
		))
	defer span.End()

	err := r.db.User.UpdateOneID(uid).SetSettings(settings).Exec(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) GetSettings(ctx context.Context, uid uuid.UUID) (map[string]interface{}, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetSettings",
		trace.WithAttributes(attribute.String("user.id", uid.String())))
	defer span.End()

	usr, err := r.db.User.Get(ctx, uid)
	if err != nil {
		recordSpanError(span, err)
		return nil, err
	}
	span.SetAttributes(attribute.Int("settings.keys.count", len(usr.Settings)))
	return usr.Settings, nil
}

func (r *UserRepository) SetOIDCIdentity(ctx context.Context, uid uuid.UUID, issuer, subject string) error {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.SetOIDCIdentity",
		trace.WithAttributes(
			attribute.String("user.id", uid.String()),
			attribute.String("oidc.issuer", issuer),
			attribute.Int("oidc.subject.length", len(subject)),
		))
	defer span.End()

	err := r.db.User.UpdateOneID(uid).SetOidcIssuer(issuer).SetOidcSubject(subject).Exec(ctx)
	recordSpanError(span, err)
	return err
}

func (r *UserRepository) GetOneOIDC(ctx context.Context, issuer, subject string) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetOneOIDC",
		trace.WithAttributes(
			attribute.String("oidc.issuer", issuer),
			attribute.Int("oidc.subject.length", len(subject)),
		))
	defer span.End()

	out, err := mapUserOutErr(r.db.User.Query().
		Where(user.OidcIssuerEQ(issuer), user.OidcSubjectEQ(subject)).
		WithGroups().
		Only(ctx))
	if err != nil {
		span.SetAttributes(
			attribute.Bool("user.found", false),
			attribute.Bool("user.lookup.not_found", ent.IsNotFound(err)),
		)
		if !ent.IsNotFound(err) {
			recordSpanError(span, err)
		}
		return out, err
	}
	span.SetAttributes(attribute.Bool("user.found", true))
	span.SetAttributes(userSpanAttrs(out)...)
	return out, nil
}

func (r *UserRepository) GetUsersByGroupID(ctx context.Context, gid uuid.UUID) ([]UserSummary, error) {
	ctx, span := entityTracer().Start(ctx, "repo.UserRepository.GetUsersByGroupID",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	out, err := mapUsersSummaryErr(r.db.User.Query().
		Where(user.HasGroupsWith(group.ID(gid))).
		WithSupervisors().
		All(ctx))
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.Int("users.count", len(out)))
	return out, nil
}
