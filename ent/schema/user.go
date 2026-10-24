package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.String("name").
			NotEmpty(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("phone_number").
			Unique().
			NotEmpty(),
		field.String("password").
			NotEmpty(),
		field.String("profile_image_path").
			Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("stakeholders", Stakeholder.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("stakeholder_groups", StakeholderGroup.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("legacies", Legacy.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("last_words", LastWord.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user_config", UserConfig.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
