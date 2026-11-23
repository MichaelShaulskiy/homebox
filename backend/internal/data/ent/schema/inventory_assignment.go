package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// InventoryAssignment tracks the quantity currently issued to a recipient.
type InventoryAssignment struct {
	ent.Schema
}

func (InventoryAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		GroupMixin{ref: "inventory_assignments", field: "group_id"},
	}
}

func (InventoryAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.Float("quantity").Default(0),
		field.UUID("entity_id", uuid.UUID{}),
		field.UUID("recipient_id", uuid.UUID{}),
		field.String("serial_number").MaxLen(255).Optional(),
		field.String("asset_tag").MaxLen(255).Optional(),
		field.String("note").MaxLen(1000).Optional(),
		field.Time("issued_at").Immutable(),
		field.Time("returned_at").Optional().Nillable(),
	}
}

func (InventoryAssignment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("group_id", "returned_at"),
	}
}

func (InventoryAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Field("entity_id").
			Ref("inventory_assignments").
			Unique().
			Required(),
		edge.From("recipient", InventoryRecipient.Type).
			Field("recipient_id").
			Ref("assignments").
			Unique().
			Required(),
	}
}
