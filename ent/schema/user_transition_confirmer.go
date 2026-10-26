package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserTransitionConfirmer holds a person who can confirm a user's transition.
type UserTransitionConfirmer struct {
	ent.Schema
}

// Fields of the UserTransitionConfirmer.
func (UserTransitionConfirmer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("name").NotEmpty(),
		field.String("email").Optional().Nillable(),
		field.String("phone").Optional().Nillable(),
		field.String("status").Default("pending"),
		field.String("invite_token_hash").Optional().Nillable().Unique(),
		field.Time("invite_expires_at").Optional().Nillable(),
		field.String("auth_token_hash").Optional().Nillable().Unique(),
		field.Time("auth_token_expires_at").Optional().Nillable(),
		field.Time("invited_at").Optional().Nillable(),
		field.Time("accepted_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Indexes of the UserTransitionConfirmer.
func (UserTransitionConfirmer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").
			Edges("owner").
			Unique().
			Annotations(entsql.IndexWhere(
				"status IN ('pending', 'accepted') AND email IS NOT NULL",
			)),
		index.Fields("phone").
			Edges("owner").
			Unique().
			Annotations(entsql.IndexWhere(
				"status IN ('pending', 'accepted') AND phone IS NOT NULL",
			)),
	}
}

// Edges of the UserTransitionConfirmer.
func (UserTransitionConfirmer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("transition_confirmers").
			Unique().
			Required().
			Immutable(),
		edge.To("stakeholder", Stakeholder.Type).
			Unique(),
		edge.To("confirmations", UserTransitionConfirmation.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
