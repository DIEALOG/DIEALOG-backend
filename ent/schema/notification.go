package schema

import (
 "time"

 "entgo.io/ent"
 "entgo.io/ent/schema/field"
 "github.com/google/uuid"
)

// Notification holds the schema definition for the Notification entity.
type Notification struct {
 ent.Schema
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
 return []ent.Field{
  field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("id"),
  field.UUID("receiver_id", uuid.UUID{}).Immutable(),
  field.String("type"),
  field.String("message"),
  field.String("metadata"),
  field.Bool("is_read"),
  field.Time("created_at").Default(time.Now).Immutable(),
 }
}

// Edges of the Notification.
func (Notification) Edges() []ent.Edge {
 return []ent.Edge{
  // No edges needed for standalone notification
 }
}