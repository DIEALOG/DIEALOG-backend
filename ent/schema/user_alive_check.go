package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserAliveCheck holds a scheduled liveness confirmation request.
type UserAliveCheck struct {
	ent.Schema
}

// Fields of the UserAliveCheck.
func (UserAliveCheck) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.Int("sequence").Positive(),
		field.Time("scheduled_at"),
		field.Time("due_at"),
		field.Time("responded_at").Optional().Nillable(),
		field.String("status").Default("scheduled"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes of the UserAliveCheck.
func (UserAliveCheck) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "sequence").Unique(),
	}
}

// Edges of the UserAliveCheck.
func (UserAliveCheck) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("alive_checks").
			Field("user_id").
			Unique().
			Required().
			Immutable(),
	}
}
