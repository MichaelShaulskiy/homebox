package repo

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/inventoryassignment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/inventoryrecipient"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/inventorytransaction"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
)

type InventoryRepository struct {
	db *ent.Client
}

type (
	InventoryRecipientCreate struct {
		FirstName  string `json:"firstName"  validate:"required,min=1,max=255"`
		LastName   string `json:"lastName"   validate:"required,min=1,max=255"`
		Email      string `json:"email"      validate:"omitempty,email,max=255"`
		Department string `json:"department" validate:"max=255"`
		IsActive   bool   `json:"isActive"`
	}

	InventoryRecipientUpdate struct {
		FirstName  string `json:"firstName"  validate:"required,min=1,max=255"`
		LastName   string `json:"lastName"   validate:"required,min=1,max=255"`
		Email      string `json:"email"      validate:"omitempty,email,max=255"`
		Department string `json:"department" validate:"max=255"`
		IsActive   bool   `json:"isActive"`
	}

	InventoryRecipientOut struct {
		ID         uuid.UUID `json:"id"`
		FirstName  string    `json:"firstName"`
		LastName   string    `json:"lastName"`
		Email      *string   `json:"email"`
		Department string    `json:"department"`
		IsActive   bool      `json:"isActive"`
		CreatedAt  time.Time `json:"createdAt"`
		UpdatedAt  time.Time `json:"updatedAt"`
	}

	InventoryRecipientQuery struct {
		Search   string `json:"search"`
		Active   *bool  `json:"active"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
	}

	InventoryIssueInput struct {
		ItemID       uuid.UUID `json:"itemId"       validate:"required"`
		RecipientID  uuid.UUID `json:"recipientId"  validate:"required"`
		Quantity     float64   `json:"quantity"     validate:"required,gt=0"`
		SerialNumber string    `json:"serialNumber" validate:"max=255"`
		AssetTag     string    `json:"assetTag"     validate:"max=255"`
		Note         string    `json:"note"         validate:"max=2000"`
	}

	InventoryReturnInput struct {
		AssignmentID uuid.UUID `json:"assignmentId" validate:"required"`
		Quantity     float64   `json:"quantity"     validate:"required,gt=0"`
		Status       string    `json:"status"       validate:"required,oneof=OK DEFECTIVE WRITTEN_OFF"`
		ReasonOrNote string    `json:"reasonOrNote" validate:"max=2000"`
	}

	InventoryStockAddInput struct {
		ItemID       uuid.UUID `json:"itemId"       validate:"required"`
		Quantity     float64   `json:"quantity"     validate:"required,gt=0"`
		ReasonOrNote string    `json:"reasonOrNote" validate:"max=2000"`
	}

	InventoryStockWriteOffInput struct {
		ItemID       uuid.UUID `json:"itemId"       validate:"required"`
		Quantity     float64   `json:"quantity"     validate:"required,gt=0"`
		ReasonOrNote string    `json:"reasonOrNote" validate:"max=2000"`
	}

	InventoryAssignmentOut struct {
		ID            uuid.UUID  `json:"id"`
		ItemID        uuid.UUID  `json:"itemId"`
		ItemName      string     `json:"itemName"`
		RecipientID   uuid.UUID  `json:"recipientId"`
		RecipientName string     `json:"recipientName"`
		Quantity      float64    `json:"quantity"`
		SerialNumber  string     `json:"serialNumber"`
		AssetTag      string     `json:"assetTag"`
		Note          string     `json:"note"`
		IssuedAt      time.Time  `json:"issuedAt"`
		ReturnedAt    *time.Time `json:"returnedAt"`
	}

	InventoryTransactionOut struct {
		ID                   uuid.UUID `json:"id"`
		Timestamp            time.Time `json:"timestamp"`
		ActionType           string    `json:"actionType"`
		ItemID               uuid.UUID `json:"itemId"`
		ItemName             string    `json:"itemName"`
		Quantity             float64   `json:"quantity"`
		PerformedByAdminID   uuid.UUID `json:"performedByAdminId"`
		PerformedByAdminName string    `json:"performedByAdminName"`
		TargetUserID         uuid.UUID `json:"targetUserId"`
		TargetUserName       string    `json:"targetUserName"`
		ReasonOrNote         string    `json:"reasonOrNote"`
		SerialNumber         string    `json:"serialNumber"`
		AssetTag             string    `json:"assetTag"`
	}

	InventoryTransactionQuery struct {
		Search      string    `json:"search"`
		ItemID      uuid.UUID `json:"itemId"`
		RecipientID uuid.UUID `json:"recipientId"`
		AdminID     uuid.UUID `json:"adminId"`
		ActionTypes []string  `json:"actionTypes"`
		StartTime   time.Time `json:"startTime"`
		EndTime     time.Time `json:"endTime"`
		Page        int       `json:"page"`
		PageSize    int       `json:"pageSize"`
	}
)

func mapInventoryRecipient(r *ent.InventoryRecipient) InventoryRecipientOut {
	return InventoryRecipientOut{
		ID: r.ID, FirstName: r.FirstName, LastName: r.LastName,
		Email: r.Email, Department: r.Department, IsActive: r.IsActive,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (r *InventoryRepository) CreateRecipient(ctx context.Context, gid uuid.UUID, data InventoryRecipientCreate) (InventoryRecipientOut, error) {
	email := strings.TrimSpace(strings.ToLower(data.Email))
	builder := r.db.InventoryRecipient.Create().
		SetGroupID(gid).
		SetFirstName(strings.TrimSpace(data.FirstName)).
		SetLastName(strings.TrimSpace(data.LastName)).
		SetDepartment(strings.TrimSpace(data.Department)).
		SetIsActive(data.IsActive)
	if email != "" {
		builder.SetEmail(email)
	}
	result, err := builder.Save(ctx)
	if err != nil {
		return InventoryRecipientOut{}, err
	}
	return mapInventoryRecipient(result), nil
}

func (r *InventoryRepository) UpdateRecipient(ctx context.Context, gid, id uuid.UUID, data InventoryRecipientUpdate) (InventoryRecipientOut, error) {
	email := strings.TrimSpace(strings.ToLower(data.Email))
	builder := r.db.InventoryRecipient.UpdateOneID(id).
		Where(inventoryrecipient.HasGroupWith(group.ID(gid))).
		SetFirstName(strings.TrimSpace(data.FirstName)).
		SetLastName(strings.TrimSpace(data.LastName)).
		SetDepartment(strings.TrimSpace(data.Department)).
		SetIsActive(data.IsActive).
		ClearEmail()
	if email != "" {
		builder.SetEmail(email)
	}
	result, err := builder.Save(ctx)
	if err != nil {
		return InventoryRecipientOut{}, err
	}
	return mapInventoryRecipient(result), nil
}

func (r *InventoryRepository) DeactivateRecipient(ctx context.Context, gid, id uuid.UUID) (InventoryRecipientOut, error) {
	result, err := r.db.InventoryRecipient.UpdateOneID(id).
		Where(inventoryrecipient.HasGroupWith(group.ID(gid))).
		SetIsActive(false).
		Save(ctx)
	if err != nil {
		return InventoryRecipientOut{}, err
	}
	return mapInventoryRecipient(result), nil
}

func (r *InventoryRepository) GetRecipient(ctx context.Context, gid, id uuid.UUID) (InventoryRecipientOut, error) {
	result, err := r.db.InventoryRecipient.Query().
		Where(inventoryrecipient.ID(id), inventoryrecipient.HasGroupWith(group.ID(gid))).
		Only(ctx)
	if err != nil {
		return InventoryRecipientOut{}, err
	}
	return mapInventoryRecipient(result), nil
}

func (r *InventoryRepository) GetRecipients(ctx context.Context, gid uuid.UUID, query InventoryRecipientQuery) ([]InventoryRecipientOut, error) {
	q := r.db.InventoryRecipient.Query().Where(inventoryrecipient.HasGroupWith(group.ID(gid)))
	if query.Active != nil {
		q = q.Where(inventoryrecipient.IsActive(*query.Active))
	}
	if strings.TrimSpace(query.Search) != "" {
		search := strings.TrimSpace(query.Search)
		q = q.Where(inventoryrecipient.Or(
			inventoryrecipient.FirstNameContainsFold(search),
			inventoryrecipient.LastNameContainsFold(search),
			inventoryrecipient.EmailContainsFold(search),
			inventoryrecipient.DepartmentContainsFold(search),
		))
	}
	if query.PageSize > 0 {
		page := query.Page
		if page < 1 {
			page = 1
		}
		q = q.Offset((page - 1) * query.PageSize).Limit(query.PageSize)
	}
	rows, err := q.Order(ent.Asc(inventoryrecipient.FieldLastName), ent.Asc(inventoryrecipient.FieldFirstName)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]InventoryRecipientOut, len(rows))
	for i, row := range rows {
		out[i] = mapInventoryRecipient(row)
	}
	return out, nil
}

func validateMovementQuantity(quantity float64) error {
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) || quantity <= 0 {
		return fmt.Errorf("quantity must be a finite number greater than zero")
	}
	return nil
}

func (r *InventoryRepository) Issue(ctx context.Context, gid uuid.UUID, data InventoryIssueInput, adminID uuid.UUID, adminName string) (InventoryAssignmentOut, error) {
	if err := validateMovementQuantity(data.Quantity); err != nil {
		return InventoryAssignmentOut{}, err
	}
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return InventoryAssignmentOut{}, err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	item, err := tx.Entity.Query().Where(entity.ID(data.ItemID), entity.HasGroupWith(group.ID(gid))).Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	recipient, err := tx.InventoryRecipient.Query().Where(
		inventoryrecipient.ID(data.RecipientID),
		inventoryrecipient.HasGroupWith(group.ID(gid)),
	).Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	if !recipient.IsActive {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, fmt.Errorf("recipient is inactive")
	}
	if item.AvailableQuantity < data.Quantity {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, fmt.Errorf("insufficient available quantity")
	}

	now := time.Now().UTC()
	item, err = tx.Entity.UpdateOne(item).
		SetAvailableQuantity(item.AvailableQuantity - data.Quantity).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	assignment, err := tx.InventoryAssignment.Create().
		SetGroupID(gid).
		SetEntityID(item.ID).
		SetRecipientID(recipient.ID).
		SetQuantity(data.Quantity).
		SetSerialNumber(data.SerialNumber).
		SetAssetTag(data.AssetTag).
		SetNote(data.Note).
		SetIssuedAt(now).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	if _, err = tx.InventoryTransaction.Create().
		SetGroupID(gid).
		SetEntityID(item.ID).
		SetItemID(item.ID).
		SetItemName(item.Name).
		SetActionType(inventorytransaction.ActionTypeISSUED).
		SetQuantity(data.Quantity).
		SetPerformedByAdminID(adminID).
		SetPerformedByAdminName(adminName).
		SetTargetUserID(recipient.ID).
		SetTargetUserName(recipient.FirstName + " " + recipient.LastName).
		SetReasonOrNote(data.Note).
		SetSerialNumber(data.SerialNumber).
		SetAssetTag(data.AssetTag).
		SetTimestamp(now).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	if err = tx.Commit(); err != nil {
		return InventoryAssignmentOut{}, err
	}
	return InventoryAssignmentOut{
		ID: assignment.ID, ItemID: item.ID, ItemName: item.Name,
		RecipientID: recipient.ID, RecipientName: recipient.FirstName + " " + recipient.LastName,
		Quantity: assignment.Quantity, SerialNumber: assignment.SerialNumber,
		AssetTag: assignment.AssetTag, Note: assignment.Note, IssuedAt: assignment.IssuedAt,
	}, nil
}

func (r *InventoryRepository) Return(ctx context.Context, gid uuid.UUID, data InventoryReturnInput, adminID uuid.UUID, adminName string) (InventoryAssignmentOut, error) {
	if err := validateMovementQuantity(data.Quantity); err != nil {
		return InventoryAssignmentOut{}, err
	}
	status := strings.ToUpper(strings.TrimSpace(data.Status))
	if status != "OK" && status != "DEFECTIVE" && status != "WRITTEN_OFF" {
		return InventoryAssignmentOut{}, fmt.Errorf("invalid return status")
	}
	if status == "DEFECTIVE" && strings.TrimSpace(data.ReasonOrNote) == "" {
		return InventoryAssignmentOut{}, fmt.Errorf("a defect description is required")
	}

	tx, err := r.db.Tx(ctx)
	if err != nil {
		return InventoryAssignmentOut{}, err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	assignment, err := tx.InventoryAssignment.Query().
		Where(
			inventoryassignment.ID(data.AssignmentID),
			inventoryassignment.HasGroupWith(group.ID(gid)),
			inventoryassignment.ReturnedAtIsNil(),
		).
		WithEntity().
		WithRecipient().
		Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	if assignment.Quantity < data.Quantity {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, fmt.Errorf("return quantity exceeds assigned quantity")
	}
	item := assignment.Edges.Entity
	recipient := assignment.Edges.Recipient
	now := time.Now().UTC()

	switch status {
	case "OK":
		_, err = tx.Entity.UpdateOne(item).
			SetAvailableQuantity(item.AvailableQuantity + data.Quantity).
			Save(ctx)
	case "DEFECTIVE":
		_, err = tx.Entity.UpdateOne(item).
			SetDefectiveQuantity(item.DefectiveQuantity + data.Quantity).
			Save(ctx)
	case "WRITTEN_OFF":
		if item.Quantity < data.Quantity {
			err = fmt.Errorf("written-off quantity exceeds total quantity")
		} else {
			_, err = tx.Entity.UpdateOne(item).
				SetQuantity(item.Quantity - data.Quantity).
				Save(ctx)
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}

	remaining := assignment.Quantity - data.Quantity
	assignmentBuilder := tx.InventoryAssignment.UpdateOne(assignment).
		SetQuantity(remaining)
	if remaining == 0 {
		assignmentBuilder.SetReturnedAt(now)
	}
	assignment, err = assignmentBuilder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}

	var action inventorytransaction.ActionType
	switch status {
	case "OK":
		action = inventorytransaction.ActionTypeRETURNED_OK
	case "DEFECTIVE":
		action = inventorytransaction.ActionTypeRETURNED_DEFECTIVE
	case "WRITTEN_OFF":
		action = inventorytransaction.ActionTypeWRITTEN_OFF
	}
	if _, err = tx.InventoryTransaction.Create().
		SetGroupID(gid).
		SetEntityID(item.ID).
		SetItemID(item.ID).
		SetItemName(item.Name).
		SetActionType(action).
		SetQuantity(data.Quantity).
		SetPerformedByAdminID(adminID).
		SetPerformedByAdminName(adminName).
		SetTargetUserID(recipient.ID).
		SetTargetUserName(recipient.FirstName + " " + recipient.LastName).
		SetReasonOrNote(data.ReasonOrNote).
		SetSerialNumber(assignment.SerialNumber).
		SetAssetTag(assignment.AssetTag).
		SetTimestamp(now).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return InventoryAssignmentOut{}, err
	}
	if err = tx.Commit(); err != nil {
		return InventoryAssignmentOut{}, err
	}
	return InventoryAssignmentOut{
		ID: assignment.ID, ItemID: item.ID, ItemName: item.Name,
		RecipientID: recipient.ID, RecipientName: recipient.FirstName + " " + recipient.LastName,
		Quantity: remaining, SerialNumber: assignment.SerialNumber,
		AssetTag: assignment.AssetTag, Note: assignment.Note,
		IssuedAt: assignment.IssuedAt, ReturnedAt: assignment.ReturnedAt,
	}, nil
}

func (r *InventoryRepository) AddStock(ctx context.Context, gid uuid.UUID, data InventoryStockAddInput, adminID uuid.UUID, adminName string) error {
	if err := validateMovementQuantity(data.Quantity); err != nil {
		return err
	}
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	item, err := tx.Entity.Query().Where(entity.ID(data.ItemID), entity.HasGroupWith(group.ID(gid))).Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	item, err = tx.Entity.UpdateOne(item).
		SetQuantity(item.Quantity + data.Quantity).
		SetAvailableQuantity(item.AvailableQuantity + data.Quantity).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	now := time.Now().UTC()
	if _, err = tx.InventoryTransaction.Create().
		SetGroupID(gid).
		SetEntityID(item.ID).
		SetItemID(item.ID).
		SetItemName(item.Name).
		SetActionType(inventorytransaction.ActionTypeSTOCK_ADDED).
		SetQuantity(data.Quantity).
		SetPerformedByAdminID(adminID).
		SetPerformedByAdminName(adminName).
		SetReasonOrNote(data.ReasonOrNote).
		SetTimestamp(now).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *InventoryRepository) WriteOffStock(ctx context.Context, gid uuid.UUID, data InventoryStockWriteOffInput, adminID uuid.UUID, adminName string) error {
	if err := validateMovementQuantity(data.Quantity); err != nil {
		return err
	}
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	item, err := tx.Entity.Query().Where(entity.ID(data.ItemID), entity.HasGroupWith(group.ID(gid))).Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if item.Quantity < data.Quantity || item.AvailableQuantity < data.Quantity {
		_ = tx.Rollback()
		return fmt.Errorf("written-off quantity exceeds available quantity")
	}
	item, err = tx.Entity.UpdateOne(item).
		SetQuantity(item.Quantity - data.Quantity).
		SetAvailableQuantity(item.AvailableQuantity - data.Quantity).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	now := time.Now().UTC()
	if _, err = tx.InventoryTransaction.Create().
		SetGroupID(gid).
		SetEntityID(item.ID).
		SetItemID(item.ID).
		SetItemName(item.Name).
		SetActionType(inventorytransaction.ActionTypeWRITTEN_OFF).
		SetQuantity(data.Quantity).
		SetPerformedByAdminID(adminID).
		SetPerformedByAdminName(adminName).
		SetReasonOrNote(data.ReasonOrNote).
		SetTimestamp(now).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *InventoryRepository) GetAssignments(ctx context.Context, gid, itemID, recipientID uuid.UUID) ([]InventoryAssignmentOut, error) {
	predicates := []predicate.InventoryAssignment{inventoryassignment.HasGroupWith(group.ID(gid)), inventoryassignment.ReturnedAtIsNil()}
	if itemID != uuid.Nil {
		predicates = append(predicates, inventoryassignment.HasEntityWith(entity.ID(itemID)))
	}
	if recipientID != uuid.Nil {
		predicates = append(predicates, inventoryassignment.HasRecipientWith(inventoryrecipient.ID(recipientID)))
	}
	rows, err := r.db.InventoryAssignment.Query().
		Where(predicates...).
		WithEntity().
		WithRecipient().
		Order(ent.Desc(inventoryassignment.FieldIssuedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]InventoryAssignmentOut, len(rows))
	for i, row := range rows {
		out[i] = InventoryAssignmentOut{
			ID: row.ID, ItemID: row.Edges.Entity.ID, ItemName: row.Edges.Entity.Name,
			RecipientID:   row.Edges.Recipient.ID,
			RecipientName: row.Edges.Recipient.FirstName + " " + row.Edges.Recipient.LastName,
			Quantity:      row.Quantity, SerialNumber: row.SerialNumber, AssetTag: row.AssetTag,
			Note: row.Note, IssuedAt: row.IssuedAt, ReturnedAt: row.ReturnedAt,
		}
	}
	return out, nil
}

func (r *InventoryRepository) GetTransactions(ctx context.Context, gid uuid.UUID, query InventoryTransactionQuery) ([]InventoryTransactionOut, error) {
	predicates := []predicate.InventoryTransaction{inventorytransaction.HasGroupWith(group.ID(gid))}
	if query.ItemID != uuid.Nil {
		predicates = append(predicates, inventorytransaction.ItemIDEQ(query.ItemID))
	}
	if query.RecipientID != uuid.Nil {
		predicates = append(predicates, inventorytransaction.TargetUserIDEQ(query.RecipientID))
	}
	if query.AdminID != uuid.Nil {
		predicates = append(predicates, inventorytransaction.PerformedByAdminIDEQ(query.AdminID))
	}
	if len(query.ActionTypes) > 0 {
		values := make([]inventorytransaction.ActionType, 0, len(query.ActionTypes))
		for _, value := range query.ActionTypes {
			values = append(values, inventorytransaction.ActionType(value))
		}
		predicates = append(predicates, inventorytransaction.ActionTypeIn(values...))
	}
	if !query.StartTime.IsZero() {
		predicates = append(predicates, inventorytransaction.TimestampGTE(query.StartTime))
	}
	if !query.EndTime.IsZero() {
		predicates = append(predicates, inventorytransaction.TimestampLTE(query.EndTime))
	}
	if strings.TrimSpace(query.Search) != "" {
		search := strings.TrimSpace(query.Search)
		predicates = append(predicates, inventorytransaction.Or(
			inventorytransaction.ItemNameContainsFold(search),
			inventorytransaction.TargetUserNameContainsFold(search),
			inventorytransaction.PerformedByAdminNameContainsFold(search),
			inventorytransaction.ReasonOrNoteContainsFold(search),
		))
	}
	q := r.db.InventoryTransaction.Query().Where(predicates...)
	if query.PageSize > 0 {
		page := query.Page
		if page < 1 {
			page = 1
		}
		q = q.Offset((page - 1) * query.PageSize).Limit(query.PageSize)
	}
	rows, err := q.Order(ent.Desc(inventorytransaction.FieldTimestamp)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]InventoryTransactionOut, len(rows))
	for i, row := range rows {
		out[i] = InventoryTransactionOut{
			ID: row.ID, Timestamp: row.Timestamp, ActionType: row.ActionType.String(),
			ItemID: row.ItemID, ItemName: row.ItemName, Quantity: row.Quantity,
			PerformedByAdminID: row.PerformedByAdminID, PerformedByAdminName: row.PerformedByAdminName,
			TargetUserID: row.TargetUserID, TargetUserName: row.TargetUserName,
			ReasonOrNote: row.ReasonOrNote, SerialNumber: row.SerialNumber, AssetTag: row.AssetTag,
		}
	}
	return out, nil
}

func (r *InventoryRepository) MapEntityStock(ctx context.Context, gid, itemID uuid.UUID) (float64, float64, float64, error) {
	row, err := r.db.Entity.Query().
		Where(entity.ID(itemID), entity.HasGroupWith(group.ID(gid))).
		Only(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	return row.Quantity, row.AvailableQuantity, row.DefectiveQuantity, nil
}
