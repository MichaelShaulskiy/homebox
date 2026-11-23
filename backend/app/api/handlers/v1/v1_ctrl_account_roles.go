package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleAccountMembersGetAll godoc
//
//	@Summary	Get account members
//	@Tags		Accounts
//	@Produce	json
//	@Success	200	{array}	repo.UserSummary
//	@Router		/v1/account-members [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAccountMembersGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.UserSummary, error) {
		return ctrl.svc.User.GetAccountMembers(services.NewContext(r.Context()))
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleAccountSupervisorsGetAll godoc
//
//	@Summary	Get admin and superadmin accounts available as supervisors
//	@Tags		Accounts
//	@Produce	json
//	@Success	200	{array}	repo.UserSummary
//	@Router		/v1/account-supervisors [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAccountSupervisorsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.UserSummary, error) {
		return ctrl.svc.User.GetAccountSupervisors(services.NewContext(r.Context()))
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleAccountCreate godoc
//
//	@Summary	Create an admin or read-only account
//	@Tags		Accounts
//	@Accept		json
//	@Produce	json
//	@Param		payload	body	repo.UserAccountCreate	true	"account"
//	@Success	201		{object}	repo.UserOut
//	@Router		/v1/account-members [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAccountCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.UserAccountCreate) (repo.UserOut, error) {
		return ctrl.svc.User.CreateAccount(services.NewContext(r.Context()), body)
	}
	return adapters.Action(fn, http.StatusCreated)
}

// HandleAccountRoleUpdate godoc
//
//	@Summary	Change an account role
//	@Tags		Accounts
//	@Accept		json
//	@Produce	json
//	@Param		id		path	string						true	"account ID"
//	@Param		payload	body	repo.UserAccountRoleUpdate	true	"role"
//	@Success	200		{object}	repo.UserOut
//	@Router		/v1/account-members/{id}/role [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAccountRoleUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body repo.UserAccountRoleUpdate) (repo.UserOut, error) {
		return ctrl.svc.User.SetAccountRole(services.NewContext(r.Context()), id, body.AccountRole)
	}
	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleAccountSupervisorsUpdate godoc
//
//	@Summary	Replace account supervisors
//	@Tags		Accounts
//	@Accept		json
//	@Produce	json
//	@Param		id		path	string								true	"account ID"
//	@Param		payload	body	repo.UserAccountSupervisorsUpdate	true	"supervisors"
//	@Success	200		{object}	repo.UserOut
//	@Router		/v1/account-members/{id}/supervisors [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAccountSupervisorsUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body repo.UserAccountSupervisorsUpdate) (repo.UserOut, error) {
		return ctrl.svc.User.SetAccountSupervisors(services.NewContext(r.Context()), id, body.SupervisorIDs)
	}
	return adapters.ActionID("id", fn, http.StatusOK)
}
