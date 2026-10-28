package repository

import (
	"context"
	"encoding/json"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/notification"
)

type notificationRepository struct {
	client *ent.Client
}

func NewNotificationRepository(client *ent.Client) domain.NotificationRepository {
	return &notificationRepository{
		client: client,
	}
}

func (r *notificationRepository) Create(c context.Context, notification *domain.Notification) (*domain.Notification, error) {
	var metadataString string
	if notification.Metadata != nil {
		jsonBytes, err := json.Marshal(notification.Metadata)
		if err != nil {
			metadataString = ""
		} else {
			metadataString = string(jsonBytes)
		}
	}

	entity, err := r.client.Notification.Create().
		SetID(notification.ID).
		SetReceiverID(notification.ReceiverID).
		SetType(string(notification.Type)).
		SetMessage(notification.Message).
		SetMetadata(metadataString).
		SetIsRead(notification.IsRead).
		Save(c)
	if err != nil {
		return nil, err
	}

	result := toDomainNotification(entity)
	return result, nil
}

func (r *notificationRepository) FindAllByReceiverID(c context.Context, receiverID *domain.ID) ([]*domain.Notification, error) {
	query := r.client.Notification.Query().
		Where(notification.ReceiverIDEQ(*receiverID))

	entities, err := query.
		Order(ent.Desc(notification.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Notification, len(entities))
	for i, entity := range entities {
		result[i] = toDomainNotification(entity)
	}

	return result, nil
}

func (r *notificationRepository) FindByID(c context.Context, id *domain.ID) (*domain.Notification, error) {
	entity, err := r.client.Notification.Query().
		Where(notification.IDEQ(*id)).
		Only(c)

	if err != nil {
		return nil, err
	}

	result := toDomainNotification(entity)
	return result, nil
}

func (r *notificationRepository) UpdateIsReadAsTrue(c context.Context, id *domain.ID) error {
	_, err := r.client.Notification.UpdateOneID(*id).
		SetIsRead(true).
		Save(c)
	return err
}

func (r *notificationRepository) CountUnreadByReceiverID(c context.Context, receiverID *domain.ID) (int, error) {
	count, err := r.client.Notification.Query().
		Where(
			notification.ReceiverIDEQ(*receiverID),
			notification.IsReadEQ(false),
		).
		Count(c)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *notificationRepository) Delete(c context.Context, id *domain.ID) error {
	n, err := r.client.Notification.Delete().
		Where(notification.IDEQ(*id)).
		Exec(c)
	if err != nil {
		return err
	}
	if n == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}
