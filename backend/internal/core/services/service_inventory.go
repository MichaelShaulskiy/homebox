package services

import (
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

type InventoryService struct {
	repos *repo.AllRepos
}

func (s *InventoryService) CreateRecipient(ctx Context, data repo.InventoryRecipientCreate) (repo.InventoryRecipientOut, error) {
	return s.repos.Inventory.CreateRecipient(ctx, ctx.GID, data)
}

func (s *InventoryService) UpdateRecipient(ctx Context, id uuid.UUID, data repo.InventoryRecipientUpdate) (repo.InventoryRecipientOut, error) {
	return s.repos.Inventory.UpdateRecipient(ctx, ctx.GID, id, data)
}

func (s *InventoryService) DeactivateRecipient(ctx Context, id uuid.UUID) (repo.InventoryRecipientOut, error) {
	return s.repos.Inventory.DeactivateRecipient(ctx, ctx.GID, id)
}

func (s *InventoryService) GetRecipient(ctx Context, id uuid.UUID) (repo.InventoryRecipientOut, error) {
	return s.repos.Inventory.GetRecipient(ctx, ctx.GID, id)
}

func (s *InventoryService) GetRecipients(ctx Context, query repo.InventoryRecipientQuery) ([]repo.InventoryRecipientOut, error) {
	return s.repos.Inventory.GetRecipients(ctx, ctx.GID, query)
}

func (s *InventoryService) Issue(ctx Context, data repo.InventoryIssueInput) (repo.InventoryAssignmentOut, error) {
	adminName := "Unknown"
	if ctx.User != nil {
		adminName = ctx.User.Name
	}
	return s.repos.Inventory.Issue(ctx, ctx.GID, data, ctx.UID, adminName)
}

func (s *InventoryService) Return(ctx Context, data repo.InventoryReturnInput) (repo.InventoryAssignmentOut, error) {
	adminName := "Unknown"
	if ctx.User != nil {
		adminName = ctx.User.Name
	}
	return s.repos.Inventory.Return(ctx, ctx.GID, data, ctx.UID, adminName)
}

func (s *InventoryService) AddStock(ctx Context, data repo.InventoryStockAddInput) error {
	adminName := "Unknown"
	if ctx.User != nil {
		adminName = ctx.User.Name
	}
	return s.repos.Inventory.AddStock(ctx, ctx.GID, data, ctx.UID, adminName)
}

func (s *InventoryService) WriteOffStock(ctx Context, data repo.InventoryStockWriteOffInput) error {
	adminName := "Unknown"
	if ctx.User != nil {
		adminName = ctx.User.Name
	}
	return s.repos.Inventory.WriteOffStock(ctx, ctx.GID, data, ctx.UID, adminName)
}

func (s *InventoryService) GetAssignments(ctx Context, itemID, recipientID uuid.UUID) ([]repo.InventoryAssignmentOut, error) {
	return s.repos.Inventory.GetAssignments(ctx, ctx.GID, itemID, recipientID)
}

func (s *InventoryService) GetTransactions(ctx Context, query repo.InventoryTransactionQuery) ([]repo.InventoryTransactionOut, error) {
	return s.repos.Inventory.GetTransactions(ctx, ctx.GID, query)
}

func (s *InventoryService) GetRecipientTransactions(ctx Context, id uuid.UUID) ([]repo.InventoryTransactionOut, error) {
	return s.repos.Inventory.GetTransactions(ctx, ctx.GID, repo.InventoryTransactionQuery{
		RecipientID: id,
		PageSize:    100,
	})
}

func (s *InventoryService) GetItemTransactions(ctx Context, id uuid.UUID) ([]repo.InventoryTransactionOut, error) {
	return s.repos.Inventory.GetTransactions(ctx, ctx.GID, repo.InventoryTransactionQuery{
		ItemID:   id,
		PageSize: 100,
	})
}

func (s *InventoryService) GetStock(ctx Context, itemID uuid.UUID) (float64, float64, float64, error) {
	return s.repos.Inventory.MapEntityStock(ctx, ctx.GID, itemID)
}
