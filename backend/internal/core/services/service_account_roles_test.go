package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func TestAccountRoleSupervisorRules(t *testing.T) {
	adminCtx := tCtx
	adminCtx.User = &repo.UserOut{
		ID:          tUser.ID,
		AccountRole: user.AccountRoleAdmin,
		IsSuperuser: false,
	}

	adminAccount, err := tSvc.User.CreateAccount(adminCtx, repo.UserAccountCreate{
		Name:        "Account Admin",
		Email:       "account-admin@example.com",
		Password:    "reasonable-password",
		AccountRole: user.AccountRoleAdmin,
	})
	require.NoError(t, err)

	userAccount, err := tSvc.User.CreateAccount(adminCtx, repo.UserAccountCreate{
		Name:          "Read Only User",
		Email:         "read-only@example.com",
		Password:      "reasonable-password",
		AccountRole:   user.AccountRoleUser,
		SupervisorIDs: []uuid.UUID{adminAccount.ID},
	})
	require.NoError(t, err)
	require.Equal(t, user.AccountRoleUser, userAccount.AccountRole)

	_, err = tSvc.User.SetAccountSupervisors(adminCtx, userAccount.ID, nil)
	require.NoError(t, err)

	readonlyCtx := tCtx
	readonlyCtx.User = &repo.UserOut{
		ID:          userAccount.ID,
		AccountRole: user.AccountRoleUser,
		IsSuperuser: false,
	}
	_, err = tSvc.User.CreateAccount(readonlyCtx, repo.UserAccountCreate{
		Name:        "Forbidden",
		Email:       "forbidden@example.com",
		Password:    "reasonable-password",
		AccountRole: user.AccountRoleUser,
	})
	require.Error(t, err)
}
