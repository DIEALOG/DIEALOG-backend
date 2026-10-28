package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
)

type notificationUseCase struct {
	notificationRepository domain.NotificationRepository
	contextTimeout         time.Duration
}

func NewNotificationUseCase(notificationRepository domain.NotificationRepository, timeout time.Duration) domain.NotificationUseCase {
	return &notificationUseCase{
		notificationRepository: notificationRepository,
		contextTimeout:         timeout,
	}
}

func (nu *notificationUseCase) Create(c context.Context, notification *domain.Notification) (*domain.Notification, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	if notification.Message == "" {
		return nil, domain.NewBadRequestError(errors.New("message is required"))
	}

	n, err := nu.notificationRepository.Create(ctx, notification)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return n, nil
}

func (nu *notificationUseCase) ListByReceiverID(c context.Context, receiverID *domain.ID) ([]*domain.Notification, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	list, err := nu.notificationRepository.FindAllByReceiverID(ctx, receiverID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return list, nil
}

func (nu *notificationUseCase) MarkAsRead(c context.Context, id *domain.ID, receiverID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	err := nu.notificationRepository.UpdateIsReadAsTrue(ctx, id)
	if err != nil {
		return domain.NewInternalServerError(err)
	}

	return nil
}

func (nu *notificationUseCase) CountUnreadByReceiverID(c context.Context, receiverID *domain.ID) (int, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	count, err := nu.notificationRepository.CountUnreadByReceiverID(ctx, receiverID)
	if err != nil {
		return 0, domain.NewInternalServerError(err)
	}

	return count, nil
}

func (nu *notificationUseCase) Delete(c context.Context, id *domain.ID, receiverID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	err := nu.notificationRepository.Delete(ctx, id)
	if err != nil {
		return domain.NewInternalServerError(err)
	}

	return nil
}
