package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

const (
	defaultMinimumConfirmers      = 2
	defaultWaitingPeriodDays      = 7
	defaultAliveCheckIntervalDays = 90
)

type userTransitionPolicyUseCase struct {
	repository           domain.UserTransitionPolicyRepository
	userConfigRepository domain.UserConfigRepository
	contextTimeout       time.Duration
}

func NewUserTransitionPolicyUseCase(repository domain.UserTransitionPolicyRepository, userConfigRepository domain.UserConfigRepository, timeout time.Duration) domain.UserTransitionPolicyUseCase {
	return &userTransitionPolicyUseCase{
		repository:           repository,
		userConfigRepository: userConfigRepository,
		contextTimeout:       timeout,
	}
}

func (u *userTransitionPolicyUseCase) Get(c context.Context, userID *domain.ID) (*domain.UserTransitionPolicy, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	policy, err := u.repository.FindByUserID(ctx, userID)
	if err == nil {
		return policy, nil
	}
	if !ent.IsNotFound(err) {
		return nil, domain.NewInternalServerError(err)
	}

	return u.createDefault(ctx, userID)
}

func (u *userTransitionPolicyUseCase) Update(c context.Context, request *domain.UpdateUserTransitionPolicyRequest, userID *domain.ID) (*domain.UserTransitionPolicy, error) {
	if request == nil {
		return nil, domain.NewBadRequestError(errors.New("request is required"))
	}

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	policy, err := u.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if request.MinimumConfirmers != nil {
		policy.MinimumConfirmers = *request.MinimumConfirmers
	}
	if request.WaitingPeriodDays != nil {
		policy.WaitingPeriodDays = *request.WaitingPeriodDays
	}
	if request.AliveCheckIntervalDays != nil {
		policy.AliveCheckIntervalDays = *request.AliveCheckIntervalDays
	}

	if policy.MinimumConfirmers < 1 {
		return nil, domain.NewBadRequestError(errors.New("minimum_confirmers must be at least 1"))
	}
	if policy.WaitingPeriodDays < 0 {
		return nil, domain.NewBadRequestError(errors.New("waiting_period_days must not be negative"))
	}
	if policy.AliveCheckIntervalDays < 1 {
		return nil, domain.NewBadRequestError(errors.New("alive_check_interval_days must be at least 1"))
	}

	updated, err := u.repository.Update(ctx, policy, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	return updated, nil
}

func (u *userTransitionPolicyUseCase) createDefault(c context.Context, userID *domain.ID) (*domain.UserTransitionPolicy, error) {
	aliveCheckIntervalDays := defaultAliveCheckIntervalDays
	if u.userConfigRepository != nil {
		userConfig, err := u.userConfigRepository.FindByUserID(c, userID)
		if err == nil && userConfig.SurvivalCheckIntervalDays > 0 {
			aliveCheckIntervalDays = userConfig.SurvivalCheckIntervalDays
		}
	}

	policy, err := u.repository.Create(c, &domain.UserTransitionPolicy{
		MinimumConfirmers:      defaultMinimumConfirmers,
		WaitingPeriodDays:      defaultWaitingPeriodDays,
		AliveCheckIntervalDays: aliveCheckIntervalDays,
	}, userID)
	if err != nil {
		if ent.IsConstraintError(err) {
			existing, findErr := u.repository.FindByUserID(c, userID)
			if findErr == nil {
				return existing, nil
			}
		}
		return nil, domain.NewInternalServerError(err)
	}
	return policy, nil
}
