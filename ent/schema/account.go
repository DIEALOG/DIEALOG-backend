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
				"Telegram",
				"Instagram",
				"Other",
			),
		field.String("type_name").
			Optional().
			Nillable(),
		field.String("name").
			NotEmpty(),
		field.Enum("action").
			Values(
				"delete",
				"remain",
				"permission-move",
				"backup-and-delete",
				"other",
			).
			Default("remain"),
		field.String("action_message").
			Optional().
			Nillable(),
		field.String("data_location").
			Optional().
			Nillable(),
		field.String("account_id").
			Optional().
			Nillable(),
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
		edge.To("stakeholders", Stakeholder.Type),
		edge.To("stakeholder_groups", StakeholderGroup.Type),
	}
}
