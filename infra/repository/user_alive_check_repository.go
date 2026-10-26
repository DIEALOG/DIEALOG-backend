package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/useralivecheck"
)

type userAliveCheckRepository struct {
	client *ent.Client
}

func NewUserAliveCheckRepository(client *ent.Client) domain.UserAliveCheckRepository {
	return &userAliveCheckRepository{client: client}
}

func (r *userAliveCheckRepository) Create(c context.Context, check *domain.UserAliveCheck, userID *domain.ID) (*domain.UserAliveCheck, error) {
	entity, err := r.client.UserAliveCheck.Create().
		SetSequence(check.Sequence).
		SetScheduledAt(check.ScheduledAt).
		SetDueAt(check.DueAt).
		SetNillableRespondedAt(check.RespondedAt).
		SetStatus(string(check.Status)).
		SetOwnerID(*userID).
		Save(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserAliveCheck(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userAliveCheckRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.UserAliveCheck, error) {
	entities, err := r.client.UserAliveCheck.Query().
		Where(useralivecheck.HasOwnerWith(user.IDEQ(*userID))).
		WithOwner().
		Order(ent.Desc(useralivecheck.FieldSequence)).
		All(c)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.UserAliveCheck, len(entities))
	for i, entity := range entities {
		items[i] = toDomainUserAliveCheck(entity)
		items[i].UserID = *userID
	}
	return items, nil
}

func (r *userAliveCheckRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserAliveCheck, error) {
	entity, err := r.client.UserAliveCheck.Query().
		Where(
			useralivecheck.IDEQ(*id),
			useralivecheck.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserAliveCheck(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userAliveCheckRepository) FindLatestByUserID(c context.Context, userID *domain.ID) (*domain.UserAliveCheck, error) {
	entity, err := r.client.UserAliveCheck.Query().
		Where(useralivecheck.HasOwnerWith(user.IDEQ(*userID))).
		WithOwner().
		Order(ent.Desc(useralivecheck.FieldSequence)).
		First(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserAliveCheck(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userAliveCheckRepository) Update(c context.Context, check *domain.UserAliveCheck, userID *domain.ID) (*domain.UserAliveCheck, error) {
	builder := r.client.UserAliveCheck.Update().
		Where(
			useralivecheck.IDEQ(check.ID),
			useralivecheck.HasOwnerWith(user.IDEQ(*userID)),
			useralivecheck.StatusIn(
				string(domain.UserAliveCheckStatusScheduled),
				string(domain.UserAliveCheckStatusPending),
				string(domain.UserAliveCheckStatusOverdue),
			),
		).
		SetSequence(check.Sequence).
		SetScheduledAt(check.ScheduledAt).
		SetDueAt(check.DueAt).
		SetStatus(string(check.Status))

	if check.RespondedAt != nil {
		builder.SetRespondedAt(*check.RespondedAt)
	} else {
		builder.ClearRespondedAt()
	}

	n, err := builder.Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		current, findErr := r.FindByID(c, &check.ID, userID)
		if findErr != nil {
			return nil, findErr
		}
		if current.Status == check.Status ||
			(check.Status != domain.UserAliveCheckStatusResponded &&
				(current.Status == domain.UserAliveCheckStatusResponded ||
					current.Status == domain.UserAliveCheckStatusCancelled)) {
			return current, nil
		}
		return nil, domain.ErrInvalidTransitionState
	}

	return r.FindByID(c, &check.ID, userID)
}
