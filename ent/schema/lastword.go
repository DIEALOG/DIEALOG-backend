package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// LastWord holds the schema definition for the LastWord entity.
type LastWord struct {
	ent.Schema
}

// Fields of the LastWord.
func (LastWord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("title").
			NotEmpty(),
		field.String("message").
			NotEmpty(),
	}
}

// Edges of the LastWord.
func (LastWord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("receiver", Stakeholder.Type),
		edge.To("receiver_groups", StakeholderGroup.Type),
		edge.From("owner", User.Type).
			Ref("last_words").
			Unique().
			Required().
			Immutable(),
	}
}
