package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Stakeholder holds the schema definition for the Stakeholder entity.
type Stakeholder struct {
	ent.Schema
}

// Fields of the Stakeholder.
func (Stakeholder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.String("phone").
			Optional().
			Nillable(),
		field.String("email").
			Optional().
			Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the Stakeholder.
func (Stakeholder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("stakeholders").
			Unique().
			Required().
			Immutable(),
	}
}
