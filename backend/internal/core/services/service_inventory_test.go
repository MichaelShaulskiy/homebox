package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func TestInventoryIssueDefectiveReturnReplacement(t *testing.T) {
	ctx := tCtx

	entity, err := tRepos.Entities.Create(ctx, ctx.GID, repo.EntityCreate{
		Name:     "Dockingstation",
		Quantity: 2,
	})
	require.NoError(t, err)

	recipient, err := tSvc.Inventory.CreateRecipient(ctx, repo.InventoryRecipientCreate{
		FirstName: "Max",
		LastName:  "Mustermann",
		IsActive:  true,
	})
	require.NoError(t, err)

	issued, err := tSvc.Inventory.Issue(ctx, repo.InventoryIssueInput{
		ItemID: entity.ID, RecipientID: recipient.ID, Quantity: 1,
		AssetTag: "DOCK-1", Note: "Arbeitsplatz",
	})
	require.NoError(t, err)

	total, available, defective, err := tSvc.Inventory.GetStock(ctx, entity.ID)
	require.NoError(t, err)
	require.InDelta(t, 2, total, 0.0001)
	require.InDelta(t, 1, available, 0.0001)
	require.InDelta(t, 0, defective, 0.0001)

	_, err = tSvc.Inventory.Return(ctx, repo.InventoryReturnInput{
		AssignmentID: issued.ID,
		Quantity:     1,
		Status:       "DEFECTIVE",
		ReasonOrNote: "USB-C-Port defekt",
	})
	require.NoError(t, err)

	total, available, defective, err = tSvc.Inventory.GetStock(ctx, entity.ID)
	require.NoError(t, err)
	require.InDelta(t, 2, total, 0.0001)
	require.InDelta(t, 1, available, 0.0001)
	require.InDelta(t, 1, defective, 0.0001)

	replacement, err := tSvc.Inventory.Issue(ctx, repo.InventoryIssueInput{
		ItemID: entity.ID, RecipientID: recipient.ID, Quantity: 1,
		AssetTag: "DOCK-2", Note: "Ersatzgeraet",
	})
	require.NoError(t, err)
	require.NotEqual(t, issued.ID, replacement.ID)

	total, available, defective, err = tSvc.Inventory.GetStock(ctx, entity.ID)
	require.NoError(t, err)
	require.InDelta(t, 2, total, 0.0001)
	require.InDelta(t, 0, available, 0.0001)
	require.InDelta(t, 1, defective, 0.0001)

	log, err := tSvc.Inventory.GetItemTransactions(ctx, entity.ID)
	require.NoError(t, err)
	require.Len(t, log, 3)
	actions := []string{log[0].ActionType, log[1].ActionType, log[2].ActionType}
	require.ElementsMatch(t, []string{"ISSUED", "RETURNED_DEFECTIVE", "ISSUED"}, actions)
}

func TestInventoryCannotReturnMoreThanAssigned(t *testing.T) {
	ctx := tCtx
	entity, err := tRepos.Entities.Create(ctx, ctx.GID, repo.EntityCreate{Name: "Maus", Quantity: 1})
	require.NoError(t, err)
	recipient, err := tSvc.Inventory.CreateRecipient(ctx, repo.InventoryRecipientCreate{
		FirstName: "Erika", LastName: "Muster", IsActive: true,
	})
	require.NoError(t, err)
	issued, err := tSvc.Inventory.Issue(ctx, repo.InventoryIssueInput{
		ItemID: entity.ID, RecipientID: recipient.ID, Quantity: 1,
	})
	require.NoError(t, err)

	_, err = tSvc.Inventory.Return(ctx, repo.InventoryReturnInput{
		AssignmentID: issued.ID, Quantity: 2, Status: "OK",
	})
	require.Error(t, err)
}
