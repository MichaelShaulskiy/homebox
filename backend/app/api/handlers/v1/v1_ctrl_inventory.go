package v1

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleInventoryRecipientsGetAll godoc
//
//	@Summary	Get Inventory Recipients
//	@Tags		Inventory
//	@Produce	json
//	@Param		search	query		string	false	"search"
//	@Param		active	query		bool	false	"active filter"
//	@Success	200		{array}	repo.InventoryRecipientOut
//	@Router		/v1/inventory/recipients [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryRecipientsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request, query repo.InventoryRecipientQuery) ([]repo.InventoryRecipientOut, error) {
		if value := r.URL.Query().Get("active"); value != "" {
			active, err := strconv.ParseBool(value)
			if err == nil {
				query.Active = &active
			}
		}
		query.Page = queryIntOrNegativeOne(r.URL.Query().Get("page"))
		query.PageSize = queryIntOrNegativeOne(r.URL.Query().Get("pageSize"))
		return ctrl.svc.Inventory.GetRecipients(services.NewContext(r.Context()), query)
	}
	return adapters.Query(fn, http.StatusOK)
}

// HandleInventoryRecipientCreate godoc
//
//	@Summary	Create Inventory Recipient
//	@Tags		Inventory
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		repo.InventoryRecipientCreate	true	"recipient"
//	@Success	201		{object}	repo.InventoryRecipientOut
//	@Router		/v1/inventory/recipients [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryRecipientCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.InventoryRecipientCreate) (repo.InventoryRecipientOut, error) {
		return ctrl.svc.Inventory.CreateRecipient(services.NewContext(r.Context()), body)
	}
	return adapters.Action(fn, http.StatusCreated)
}

// HandleInventoryRecipientGet godoc
//
//	@Summary	Get Inventory Recipient
//	@Tags		Inventory
//	@Produce	json
//	@Param		id	path		string	true	"recipient ID"
//	@Success	200	{object}	repo.InventoryRecipientOut
//	@Router		/v1/inventory/recipients/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryRecipientGet() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (repo.InventoryRecipientOut, error) {
		return ctrl.svc.Inventory.GetRecipient(services.NewContext(r.Context()), id)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleInventoryRecipientUpdate godoc
//
//	@Summary	Update Inventory Recipient
//	@Tags		Inventory
//	@Accept		json
//	@Produce	json
//	@Param		id		path	string							true	"recipient ID"
//	@Param		payload	body	repo.InventoryRecipientUpdate	true	"recipient"
//	@Success	200		{object}	repo.InventoryRecipientOut
//	@Router		/v1/inventory/recipients/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryRecipientUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body repo.InventoryRecipientUpdate) (repo.InventoryRecipientOut, error) {
		return ctrl.svc.Inventory.UpdateRecipient(services.NewContext(r.Context()), id, body)
	}
	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleInventoryRecipientDeactivate godoc
//
//	@Summary	Deactivate Inventory Recipient
//	@Tags		Inventory
//	@Produce	json
//	@Param		id	path	string	true	"recipient ID"
//	@Success	200	{object}	repo.InventoryRecipientOut
//	@Router		/v1/inventory/recipients/{id}/deactivate [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryRecipientDeactivate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (repo.InventoryRecipientOut, error) {
		return ctrl.svc.Inventory.DeactivateRecipient(services.NewContext(r.Context()), id)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleInventoryIssue godoc
//
//	@Summary	Issue inventory
//	@Tags		Inventory
//	@Accept		json
//	@Produce	json
//	@Param		payload	body	repo.InventoryIssueInput	true	"issue"
//	@Success	200		{object}	repo.InventoryAssignmentOut
//	@Router		/v1/inventory/issue [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryIssue() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.InventoryIssueInput) (repo.InventoryAssignmentOut, error) {
		return ctrl.svc.Inventory.Issue(services.NewContext(r.Context()), body)
	}
	return adapters.Action(fn, http.StatusOK)
}

// HandleInventoryReturn godoc
//
//	@Summary	Return inventory
//	@Tags		Inventory
//	@Accept		json
//	@Produce	json
//	@Param		id		path	string							true	"assignment ID"
//	@Param		payload	body	repo.InventoryReturnInput	true	"return"
//	@Success	200		{object}	repo.InventoryAssignmentOut
//	@Router		/v1/inventory/assignments/{id}/return [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryReturn() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body repo.InventoryReturnInput) (repo.InventoryAssignmentOut, error) {
		body.AssignmentID = id
		return ctrl.svc.Inventory.Return(services.NewContext(r.Context()), body)
	}
	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleInventoryStockAdd godoc
//
//	@Summary	Add inventory stock
//	@Tags		Inventory
//	@Accept		json
//	@Produce	json
//	@Param		payload	body	repo.InventoryStockAddInput	true	"stock"
//	@Success	204
//	@Router		/v1/inventory/stock [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryStockAdd() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.InventoryStockAddInput) (any, error) {
		err := ctrl.svc.Inventory.AddStock(services.NewContext(r.Context()), body)
		return nil, err
	}
	return adapters.Action(fn, http.StatusNoContent)
}

// HandleInventoryStockWriteOff godoc
//
//	@Summary	Write off available inventory stock
//	@Tags		Inventory
//	@Accept		json
//	@Produce	json
//	@Param		payload	body	repo.InventoryStockWriteOffInput	true	"write-off"
//	@Success	204
//	@Router		/v1/inventory/write-off [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryStockWriteOff() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.InventoryStockWriteOffInput) (any, error) {
		err := ctrl.svc.Inventory.WriteOffStock(services.NewContext(r.Context()), body)
		return nil, err
	}
	return adapters.Action(fn, http.StatusNoContent)
}

// HandleInventoryAssignmentsGetAll godoc
//
//	@Summary	Get active inventory assignments
//	@Tags		Inventory
//	@Produce	json
//	@Param		itemId		query	string	false	"item ID"
//	@Param		recipientId	query	string	false	"recipient ID"
//	@Success	200			{array}	repo.InventoryAssignmentOut
//	@Router		/v1/inventory/assignments [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryAssignmentsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.InventoryAssignmentOut, error) {
		params := r.URL.Query()
		itemID, _ := uuid.Parse(params.Get("itemId"))
		recipientID, _ := uuid.Parse(params.Get("recipientId"))
		return ctrl.svc.Inventory.GetAssignments(services.NewContext(r.Context()), itemID, recipientID)
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleInventoryTransactionsGetAll godoc
//
//	@Summary	Get inventory movement log
//	@Tags		Inventory
//	@Produce	json
//	@Param		search			query	string	false	"search"
//	@Param		itemId			query	string	false	"item ID"
//	@Param		recipientId		query	string	false	"recipient ID"
//	@Param		adminId			query	string	false	"admin ID"
//	@Param		actionTypes		query	[]string	false	"action types" collectionFormat(multi)
//	@Param		startTime		query	string	false	"RFC3339 start"
//	@Param		endTime			query	string	false	"RFC3339 end"
//	@Param		page			query	int		false	"page"
//	@Param		pageSize		query	int		false	"page size"
//	@Success	200				{array}	repo.InventoryTransactionOut
//	@Router		/v1/inventory/transactions [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryTransactionsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.InventoryTransactionOut, error) {
		params := r.URL.Query()
		query := repo.InventoryTransactionQuery{
			Search:      params.Get("search"),
			ActionTypes: params["actionType"],
			Page:        queryIntOrNegativeOne(params.Get("page")),
			PageSize:    queryIntOrNegativeOne(params.Get("pageSize")),
		}
		query.ItemID, _ = uuid.Parse(params.Get("itemId"))
		query.RecipientID, _ = uuid.Parse(params.Get("recipientId"))
		query.AdminID, _ = uuid.Parse(params.Get("adminId"))
		if value := params.Get("startTime"); value != "" {
			query.StartTime, _ = time.Parse(time.RFC3339, value)
		}
		if value := params.Get("endTime"); value != "" {
			query.EndTime, _ = time.Parse(time.RFC3339, value)
		}
		return ctrl.svc.Inventory.GetTransactions(services.NewContext(r.Context()), query)
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleInventoryRecipientTransactions godoc
//
//	@Summary	Get recipient movement history
//	@Tags		Inventory
//	@Produce	json
//	@Param		id	path	string	true	"recipient ID"
//	@Success	200	{array}	repo.InventoryTransactionOut
//	@Router		/v1/inventory/recipients/{id}/transactions [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryRecipientTransactions() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) ([]repo.InventoryTransactionOut, error) {
		return ctrl.svc.Inventory.GetRecipientTransactions(services.NewContext(r.Context()), id)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleInventoryItemTransactions godoc
//
//	@Summary	Get item movement history
//	@Tags		Inventory
//	@Produce	json
//	@Param		id	path	string	true	"item ID"
//	@Success	200	{array}	repo.InventoryTransactionOut
//	@Router		/v1/inventory/items/{id}/transactions [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryItemTransactions() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) ([]repo.InventoryTransactionOut, error) {
		return ctrl.svc.Inventory.GetItemTransactions(services.NewContext(r.Context()), id)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleInventoryTransactionsCSV godoc
//
//	@Summary	Export inventory movement log as CSV
//	@Tags		Inventory
//	@Produce	text/csv
//	@Success	200	{string}	string
//	@Router		/v1/inventory/transactions/export [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleInventoryTransactionsCSV() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		rows, err := ctrl.svc.Inventory.GetTransactions(services.NewContext(r.Context()), repo.InventoryTransactionQuery{PageSize: 10000})
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=inventory-transactions.csv")
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{"timestamp", "action", "itemId", "itemName", "quantity", "recipientId", "recipientName", "adminId", "adminName", "reasonOrNote", "serialNumber", "assetTag"})
		for _, row := range rows {
			_ = writer.Write([]string{
				row.Timestamp.UTC().Format(time.RFC3339),
				row.ActionType,
				row.ItemID.String(),
				row.ItemName,
				strconv.FormatFloat(row.Quantity, 'f', -1, 64),
				row.TargetUserID.String(),
				row.TargetUserName,
				row.PerformedByAdminID.String(),
				row.PerformedByAdminName,
				row.ReasonOrNote,
				row.SerialNumber,
				row.AssetTag,
			})
		}
		writer.Flush()
		return writer.Error()
	}
}
