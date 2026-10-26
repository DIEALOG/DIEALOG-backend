package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AccountHandlingTaskAssignee snapshots one stakeholder who was asked to
// carry out an account's post-transition action.
type AccountHandlingTaskAssignee struct {
	ent.Schema
}

// Fields of the AccountHandlingTaskAssignee.
func (AccountHandlingTaskAssignee) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
		field.UUID("task_id", uuid.UUID{}).Immutable(),
		field.UUID("stakeholder_id", uuid.UUID{}).Immutable(),
		field.String("name").NotEmpty(),
		field.String("email").Optional().Nillable(),
		field.String("phone").Optional().Nillable(),
		field.String("status").Default("pending"),
		field.String("request_token_hash").Optional().Nillable().Unique(),
		field.Time("request_token_expires_at").Optional().Nillable(),
		field.Time("requested_at").Optional().Nillable(),
		field.Time("responded_at").Optional().Nillable(),
		field.String("response_message").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Indexes of the AccountHandlingTaskAssignee.
func (AccountHandlingTaskAssignee) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id", "stakeholder_id").Unique(),
	}
}

// Edges of the AccountHandlingTaskAssignee.
func (AccountHandlingTaskAssignee) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", AccountHandlingTask.Type).
			Ref("assignees").
			Field("task_id").
			Unique().
			Required().
			Immutable(),
	}
}
