package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

const aliveCheckResponseWindow = 7 * 24 * time.Hour

type userAliveCheckUseCase struct {
	repository     domain.UserAliveCheckRepository
	policyUseCase  domain.UserTransitionPolicyUseCase
	contextTimeout time.Duration
}

func NewUserAliveCheckUseCase(repository domain.UserAliveCheckRepository, policyUseCase domain.UserTransitionPolicyUseCase, timeout time.Duration) domain.UserAliveCheckUseCase {
	return &userAliveCheckUseCase{
		repository:     repository,
		policyUseCase:  policyUseCase,
		contextTimeout: timeout,
	}
}

func (u *userAliveCheckUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.UserAliveCheck, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	policy, err := u.policyUseCase.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := u.repository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	now := time.Now()
	for _, item := range items {
		nextStatus := aliveCheckStatus(item, now)
		if nextStatus == item.Status {
			continue
		}
		item.Status = nextStatus
		updated, updateErr := u.repository.Update(ctx, item, userID)
		if updateErr != nil {
			return nil, domain.NewInternalServerError(updateErr)
		}
		*item = *updated
	}

	var latest *domain.UserAliveCheck
	hasOpenCheck := false
	for _, item := range items {
		if latest == nil || item.Sequence > latest.Sequence {
			latest = item
		}
		if item.Status == domain.UserAliveCheckStatusScheduled ||
			item.Status == domain.UserAliveCheckStatusPending {
			hasOpenCheck = true
		}
	}

	// Self-heal a missed next check after a response/cancellation. An overdue
	// check must still be answered before another one is created.
	if !hasOpenCheck && (latest == nil || latest.Status != domain.UserAliveCheckStatusOverdue) {
		created, createErr := u.createNext(ctx, userID, policy, latest, now)
		if createErr != nil {
			return nil, createErr
		}
		if latest == nil || created.Sequence > latest.Sequence {
			items = append([]*domain.UserAliveCheck{created}, items...)
		}
	}

	return items, nil
}

func (u *userAliveCheckUseCase) ScheduleNext(c context.Context, userID *domain.ID) (*domain.UserAliveCheck, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	policy, err := u.policyUseCase.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	latest, err := u.repository.FindLatestByUserID(ctx, userID)
	if err != nil && !ent.IsNotFound(err) {
		return nil, domain.NewInternalServerError(err)
	}
	if err == nil {
		currentStatus := aliveCheckStatus(latest, now)
		if currentStatus != latest.Status {
			latest.Status = currentStatus
			updated, updateErr := u.repository.Update(ctx, latest, userID)
			if updateErr != nil {
				return nil, domain.NewInternalServerError(updateErr)
			}
			latest = updated
		}
		if latest.Status == domain.UserAliveCheckStatusScheduled ||
			latest.Status == domain.UserAliveCheckStatusPending {
			return latest, nil
		}
		if latest.Status == domain.UserAliveCheckStatusOverdue {
			return nil, domain.NewBadRequestError(errors.New("respond to the overdue alive check before scheduling another"))
		}
	}

	return u.createNext(ctx, userID, policy, latest, now)
}

func (u *userAliveCheckUseCase) Respond(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserAliveCheck, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	if item.Status == domain.UserAliveCheckStatusResponded ||
		item.Status == domain.UserAliveCheckStatusCancelled {
		return nil, domain.NewBadRequestError(errors.New("alive check has already been closed"))
	}

	// Fetch the policy before mutating anything so a policy failure cannot
	// leave the response persisted without a follow-up check.
	policy, err := u.policyUseCase.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	item.Status = domain.UserAliveCheckStatusResponded
	item.RespondedAt = &now
	updated, err := u.repository.Update(ctx, item, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return nil, domain.NewBadRequestError(errors.New("alive check has already been closed"))
		}
		return nil, domain.NewInternalServerError(err)
	}

	if _, err := u.createNext(ctx, userID, policy, updated, now); err != nil {
		return nil, err
	}

	return updated, nil
}

func (u *userAliveCheckUseCase) createNext(c context.Context, userID *domain.ID, policy *domain.UserTransitionPolicy, latest *domain.UserAliveCheck, now time.Time) (*domain.UserAliveCheck, error) {
	sequence := 1
	if latest != nil {
		if latest.Status == domain.UserAliveCheckStatusOverdue {
			return nil, domain.NewBadRequestError(errors.New("respond to the overdue alive check before scheduling another"))
		}
		sequence = latest.Sequence + 1
	}

	scheduledAt := now.Add(time.Duration(policy.AliveCheckIntervalDays) * 24 * time.Hour)
	dueAt := scheduledAt.Add(aliveCheckResponseWindow)
	created, err := u.repository.Create(c, &domain.UserAliveCheck{
		Sequence:    sequence,
		ScheduledAt: scheduledAt,
		DueAt:       dueAt,
		Status:      domain.UserAliveCheckStatusScheduled,
	}, userID)
	if err != nil {
		if ent.IsConstraintError(err) {
			existing, findErr := u.repository.FindLatestByUserID(c, userID)
			if findErr == nil {
				return existing, nil
			}
		}
		return nil, domain.NewInternalServerError(err)
	}
	return created, nil
}

func aliveCheckStatus(item *domain.UserAliveCheck, now time.Time) domain.UserAliveCheckStatus {
	if item.Status == domain.UserAliveCheckStatusResponded ||
		item.Status == domain.UserAliveCheckStatusCancelled {
		return item.Status
	}
	if !now.Before(item.DueAt) {
		return domain.UserAliveCheckStatusOverdue
	}
	if !now.Before(item.ScheduledAt) {
		return domain.UserAliveCheckStatusPending
	}
	return domain.UserAliveCheckStatusScheduled
}
