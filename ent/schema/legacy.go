package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Legacy holds the schema definition for the Legacy entity.
type Legacy struct {
	ent.Schema
}

// Fields of the Legacy.
func (Legacy) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the Legacy.
func (Legacy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("legacies").
			Unique().
			Required().
			Immutable(),
		edge.To("accounts", Account.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
