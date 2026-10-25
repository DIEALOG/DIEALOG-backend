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

// AccountHandlingTask holds the post-transition action for one account.
type AccountHandlingTask struct {
	ent.Schema
}

// Fields of the AccountHandlingTask.
func (AccountHandlingTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.UUID("transition_case_id", uuid.UUID{}).Immutable(),
		field.UUID("account_id", uuid.UUID{}).Immutable(),
		field.String("action_snapshot").NotEmpty(),
		field.String("action_message_snapshot").Optional().Nillable(),
		field.String("status").Default("pending"),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.String("failure_reason").Optional().Nillable(),
		field.Int("request_count").NonNegative().Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Indexes of the AccountHandlingTask.
func (AccountHandlingTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("transition_case_id", "account_id").Unique(),
	}
}

// Edges of the AccountHandlingTask.
func (AccountHandlingTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("transition_case", UserTransitionCase.Type).
			Ref("tasks").
			Field("transition_case_id").
			Unique().
			Required().
			Immutable(),
		edge.From("account", Account.Type).
			Ref("handling_tasks").
			Field("account_id").
			Unique().
			Required().
			Immutable(),
		edge.To("events", AccountHandlingEvent.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("assignees", AccountHandlingTaskAssignee.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
