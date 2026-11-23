package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// InventoryRecipient represents a person who may receive inventory without
// requiring a Homebox system login.
type InventoryRecipient struct {
	ent.Schema
}

func (InventoryRecipient) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		GroupMixin{ref: "inventory_recipients", field: "group_id"},
	}
}

func (InventoryRecipient) Fields() []ent.Field {
	return []ent.Field{
		field.String("first_name").MaxLen(255).NotEmpty(),
		field.String("last_name").MaxLen(255).NotEmpty(),
		field.String("email").MaxLen(255).Optional().Nillable(),
		field.String("department").MaxLen(255).Optional(),
		field.Bool("is_active").Default(true),
	}
}

func (InventoryRecipient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("group_id", "is_active"),
		index.Fields("group_id", "email").Unique(),
		index.Fields("last_name"),
	}
}

func (InventoryRecipient) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("assignments", InventoryAssignment.Type),
		edge.To("inventory_transactions", InventoryTransaction.Type),
	}
}
