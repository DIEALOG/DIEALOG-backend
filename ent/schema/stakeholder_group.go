package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// StakeholderGroup holds the schema definition for the StakeholderGroup entity.
type StakeholderGroup struct {
	ent.Schema
}

// Fields of the StakeholderGroup.
func (StakeholderGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the StakeholderGroup.
func (StakeholderGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("stakeholder_groups").
			Unique().
			Required().
			Immutable(),
		edge.To("stakeholders", Stakeholder.Type),
	}
}
