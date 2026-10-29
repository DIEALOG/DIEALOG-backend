package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID         ID               `json:"id"`
	ReceiverID ID               `json:"receiver_id"`
	Type       NotificationType `json:"type"`
	Message    string           `json:"message"`
	Metadata   map[string]any   `json:"metadata"  swaggertype:"object"`
	IsRead     bool             `json:"is_read"`
	CreatedAt  time.Time        `json:"created_at"`
}

type NotificationCountResponse struct {
	Count int `json:"count"`
}

type NotificationType string

const (
	NotificationTypeTransition NotificationType = "transition"
	NotificationTypeDiealog    NotificationType = "legacy"
	NotificationTypeInvite     NotificationType = "invite"
)

type NotificationRepository interface {
	Create(c context.Context, notification *Notification) (*Notification, error)
	FindAllByReceiverID(c context.Context, receiverID *ID) ([]*Notification, error)
	FindByID(c context.Context, id *ID) (*Notification, error)
	UpdateIsReadAsTrue(c context.Context, id *ID) error
	CountUnreadByReceiverID(c context.Context, receiverID *ID) (int, error)
	Delete(c context.Context, id *ID) error
}

type NotificationUseCase interface {
	Create(c context.Context, notification *Notification) (*Notification, error)
	ListByReceiverID(c context.Context, receiverID *ID) ([]*Notification, error)
	MarkAsRead(c context.Context, id *ID, receiverID *ID) error
	CountUnreadByReceiverID(c context.Context, receiverID *ID) (int, error)
	Delete(c context.Context, id *ID, receiverID *ID) error
}
