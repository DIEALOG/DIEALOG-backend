package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserTransitionCase holds one post-transition confirmation process.
type UserTransitionCase struct {
	ent.Schema
}

// Fields of the UserTransitionCase.
func (UserTransitionCase) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("active_slot").Optional().Nillable().Unique(),
		field.String("status").Default("pending_confirmation"),
		field.Int("minimum_confirmers_snapshot").Positive(),
		field.Int("waiting_period_days_snapshot").NonNegative(),
		field.Time("requested_at").Default(time.Now),
		field.Time("waiting_started_at").Optional().Nillable(),
		field.Time("eligible_at").Optional().Nillable(),
		field.Time("confirmed_at").Optional().Nillable(),
		field.Time("cancelled_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the UserTransitionCase.
func (UserTransitionCase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("transition_cases").
			Unique().
			Required().
			Immutable(),
		edge.To("confirmations", UserTransitionConfirmation.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tasks", AccountHandlingTask.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
