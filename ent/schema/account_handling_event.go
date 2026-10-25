package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// AccountHandlingEvent holds an append-only task history entry.
type AccountHandlingEvent struct {
	ent.Schema
}

// Fields of the AccountHandlingEvent.
func (AccountHandlingEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("event_type").NotEmpty(),
		field.String("message").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the AccountHandlingEvent.
func (AccountHandlingEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", AccountHandlingTask.Type).
			Ref("events").
			Unique().
			Required().
			Immutable(),
	}
}
