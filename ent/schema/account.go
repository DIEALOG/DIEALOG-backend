package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Account holds the schema definition for the Account entity.
type Account struct {
	ent.Schema
}

// Fields of the Account.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.Enum("type").
			Values(
				"Google",
				"Facebook",
				"Twitter",
				"Apple",
				"Discord",
				"Microsoft",
				"KakaoTalk",
				"GitHub",
				"Other",
			),
		field.String("account_type_other").
			Optional().
			Nillable(),
		field.String("account_id").
			NotEmpty(),
		field.String("account_password").
			NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("legacy", Legacy.Type).
			Ref("accounts").
			Unique().
			Required().
			Immutable(),
	}
}
