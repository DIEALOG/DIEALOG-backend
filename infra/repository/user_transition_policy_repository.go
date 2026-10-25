package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitionpolicy"
)

type userTransitionPolicyRepository struct {
	client *ent.Client
}

func NewUserTransitionPolicyRepository(client *ent.Client) domain.UserTransitionPolicyRepository {
	return &userTransitionPolicyRepository{client: client}
}

func (r *userTransitionPolicyRepository) FindByUserID(c context.Context, userID *domain.ID) (*domain.UserTransitionPolicy, error) {
	entity, err := r.client.UserTransitionPolicy.Query().
		Where(usertransitionpolicy.HasOwnerWith(user.IDEQ(*userID))).
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionPolicy(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userTransitionPolicyRepository) Create(c context.Context, policy *domain.UserTransitionPolicy, userID *domain.ID) (*domain.UserTransitionPolicy, error) {
	entity, err := r.client.UserTransitionPolicy.Create().
		SetMinimumConfirmers(policy.MinimumConfirmers).
		SetWaitingPeriodDays(policy.WaitingPeriodDays).
		SetAliveCheckIntervalDays(policy.AliveCheckIntervalDays).
		SetOwnerID(*userID).
		Save(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionPolicy(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userTransitionPolicyRepository) Update(c context.Context, policy *domain.UserTransitionPolicy, userID *domain.ID) (*domain.UserTransitionPolicy, error) {
	exists, err := r.client.UserTransitionPolicy.Query().
		Where(usertransitionpolicy.HasOwnerWith(user.IDEQ(*userID))).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	_, err = r.client.UserTransitionPolicy.Update().
		Where(usertransitionpolicy.HasOwnerWith(user.IDEQ(*userID))).
		SetMinimumConfirmers(policy.MinimumConfirmers).
		SetWaitingPeriodDays(policy.WaitingPeriodDays).
		SetAliveCheckIntervalDays(policy.AliveCheckIntervalDays).
		Save(c)
	if err != nil {
		return nil, err
	}

	entity, err := r.client.UserTransitionPolicy.Query().
		Where(usertransitionpolicy.HasOwnerWith(user.IDEQ(*userID))).
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionPolicy(entity)
	item.UserID = *userID
	return item, nil
}
