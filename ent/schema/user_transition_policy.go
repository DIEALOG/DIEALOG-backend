package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserTransitionPolicy holds the per-user post-transition policy.
type UserTransitionPolicy struct {
	ent.Schema
}

// Fields of the UserTransitionPolicy.
func (UserTransitionPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.Int("minimum_confirmers").Positive().Default(2),
		field.Int("waiting_period_days").NonNegative().Default(7),
		field.Int("alive_check_interval_days").Positive().Default(90),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the UserTransitionPolicy.
func (UserTransitionPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("transition_policy").
			Unique().
			Required().
			Immutable(),
	}
}
