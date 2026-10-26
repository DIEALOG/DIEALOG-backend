package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserTransitionConfirmation holds one confirmer's decision for a case.
type UserTransitionConfirmation struct {
	ent.Schema
}

// Fields of the UserTransitionConfirmation.
func (UserTransitionConfirmation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.UUID("transition_case_id", uuid.UUID{}).Immutable(),
		field.UUID("transition_confirmer_id", uuid.UUID{}).Immutable(),
		field.String("decision"),
		field.String("comment").Optional().Nillable(),
		field.Time("responded_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes of the UserTransitionConfirmation.
func (UserTransitionConfirmation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("transition_case_id", "transition_confirmer_id").Unique(),
	}
}

// Edges of the UserTransitionConfirmation.
func (UserTransitionConfirmation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("transition_case", UserTransitionCase.Type).
			Ref("confirmations").
			Field("transition_case_id").
			Unique().
			Required().
			Immutable(),
		edge.From("transition_confirmer", UserTransitionConfirmer.Type).
			Ref("confirmations").
			Field("transition_confirmer_id").
			Unique().
			Required().
			Immutable(),
	}
}
