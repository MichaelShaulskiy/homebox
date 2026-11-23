package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// InventoryTransaction is an append-only movement record. It is never updated
// or deleted by application endpoints.
type InventoryTransaction struct {
	ent.Schema
}

func (InventoryTransaction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		GroupMixin{ref: "inventory_transactions", field: "group_id"},
	}
}

func (InventoryTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.Time("timestamp").
			Immutable().
			Default(time.Now),
		field.Enum("action_type").
			Immutable().
			Values("ISSUED", "RETURNED_OK", "RETURNED_DEFECTIVE", "WRITTEN_OFF", "STOCK_ADDED"),
		field.UUID("item_id", uuid.UUID{}).Immutable(),
		field.String("item_name").MaxLen(255).Immutable(),
		field.Float("quantity").Immutable().Positive(),
		field.UUID("performed_by_admin_id", uuid.UUID{}).Immutable(),
		field.String("performed_by_admin_name").MaxLen(255).Immutable(),
		field.UUID("target_user_id", uuid.UUID{}).Immutable().Optional(),
		field.String("target_user_name").MaxLen(255).Immutable().Optional(),
		field.String("reason_or_note").MaxLen(2000).Immutable().Optional(),
		field.String("serial_number").MaxLen(255).Immutable().Optional(),
		field.String("asset_tag").MaxLen(255).Immutable().Optional(),
	}
}

func (InventoryTransaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("timestamp"),
		index.Fields("group_id", "action_type"),
		index.Fields("item_id"),
		index.Fields("target_user_id"),
		index.Fields("performed_by_admin_id"),
	}
}

func (InventoryTransaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Field("item_id").
			Ref("inventory_transactions").
			Unique().
			Immutable().
			Required(),
		edge.From("recipient", InventoryRecipient.Type).
			Field("target_user_id").
			Ref("inventory_transactions").
			Unique().
			Immutable(),
	}
}
