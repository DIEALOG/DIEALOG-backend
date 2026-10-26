package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/accounthandlingevent"
	"github.com/DIEALOG/DIEALOG-backend/ent/accounthandlingtask"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitioncase"
)

type accountHandlingEventRepository struct {
	client *ent.Client
}

func NewAccountHandlingEventRepository(client *ent.Client) domain.AccountHandlingEventRepository {
	return &accountHandlingEventRepository{client: client}
}

func (r *accountHandlingEventRepository) Create(c context.Context, event *domain.AccountHandlingEvent, userID *domain.ID) (*domain.AccountHandlingEvent, error) {
	exists, err := r.client.AccountHandlingTask.Query().
		Where(
			accounthandlingtask.IDEQ(event.TaskID),
			accounthandlingtask.HasTransitionCaseWith(
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	entity, err := r.client.AccountHandlingEvent.Create().
		SetTaskID(event.TaskID).
		SetEventType(string(event.EventType)).
		SetNillableMessage(event.Message).
		Save(c)
	if err != nil {
		return nil, err
	}

	item := toDomainAccountHandlingEvent(entity)
	item.TaskID = event.TaskID
	return item, nil
}

func (r *accountHandlingEventRepository) FindByTaskID(c context.Context, taskID *domain.ID, userID *domain.ID) ([]*domain.AccountHandlingEvent, error) {
	entities, err := r.client.AccountHandlingEvent.Query().
		Where(
			accounthandlingevent.HasTaskWith(
				accounthandlingtask.IDEQ(*taskID),
				accounthandlingtask.HasTransitionCaseWith(
					usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
				),
			),
		).
		WithTask().
		Order(ent.Asc(accounthandlingevent.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.AccountHandlingEvent, len(entities))
	for i, entity := range entities {
		items[i] = toDomainAccountHandlingEvent(entity)
	}
	return items, nil
}
