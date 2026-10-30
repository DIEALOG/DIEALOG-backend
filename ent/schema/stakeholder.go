package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.String("relationship").
			Default(""),
		// stakeholder_id holds the DIEALOG user this contact belongs to, when the
		// registered email/phone matches an existing account. It is resolved on
		// registration; a stakeholder does not have to be a member.
		field.UUID("stakeholder_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes of the Stakeholder.
func (Stakeholder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("stakeholder_id"),
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
		edge.From("groups", StakeholderGroup.Type).
			Ref("stakeholders"),
		edge.From("accounts", Account.Type).
			Ref("stakeholders"),
		edge.From("last_words", LastWord.Type).
			Ref("receiver"),
	}
}
