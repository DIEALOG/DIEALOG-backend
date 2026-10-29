package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type userTransitionCaseUseCase struct {
	repository             domain.UserTransitionCaseRepository
	confirmationRepository domain.UserTransitionConfirmationRepository
	policyUseCase          domain.UserTransitionPolicyUseCase
	confirmerRepository    domain.UserTransitionConfirmerRepository
	stakeholderRepository  domain.StakeholderRepository
	userRepository         domain.UserRepository
	notificationUseCase    domain.NotificationUseCase
	contextTimeout         time.Duration
}

func NewUserTransitionCaseUseCase(
	repository domain.UserTransitionCaseRepository,
	confirmationRepository domain.UserTransitionConfirmationRepository,
	policyUseCase domain.UserTransitionPolicyUseCase,
	confirmerRepository domain.UserTransitionConfirmerRepository,
	stakeholderRepository domain.StakeholderRepository,
	userRepository domain.UserRepository,
	notificationUseCase domain.NotificationUseCase,
	timeout time.Duration,
) domain.UserTransitionCaseUseCase {
	return &userTransitionCaseUseCase{
		repository:             repository,
		confirmationRepository: confirmationRepository,
		policyUseCase:          policyUseCase,
		confirmerRepository:    confirmerRepository,
		stakeholderRepository:  stakeholderRepository,
		userRepository:         userRepository,
		notificationUseCase:    notificationUseCase,
		contextTimeout:         timeout,
	}
}

func (u *userTransitionCaseUseCase) Start(c context.Context, confirmer *domain.UserTransitionConfirmer) (*domain.UserTransitionCase, error) {
	if confirmer == nil {
		return nil, domain.NewUnauthorizedError(errors.New("confirmer is required"))
	}
	if confirmer.Status != domain.UserTransitionConfirmerStatusAccepted {
		return nil, domain.NewForbiddenError(domain.ErrConfirmerNotAccepted)
	}

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	_, err := u.repository.FindActiveByUserID(ctx, &confirmer.UserID)
	if err == nil {
		return nil, domain.NewBadRequestError(domain.ErrActiveTransitionCaseExists)
	}
	if !ent.IsNotFound(err) {
		return nil, domain.NewInternalServerError(err)
	}

	policy, err := u.policyUseCase.Get(ctx, &confirmer.UserID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	created, err := u.repository.Create(ctx, &domain.UserTransitionCase{
		Status:                    domain.UserTransitionCaseStatusPending,
		MinimumConfirmersSnapshot: policy.MinimumConfirmers,
		WaitingPeriodDaysSnapshot: policy.WaitingPeriodDays,
		RequestedAt:               now,
	}, &confirmer.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrActiveTransitionCaseExists) {
			return nil, domain.NewBadRequestError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	if u.notificationUseCase != nil && u.userRepository != nil && u.confirmerRepository != nil {
		targetUserName := "지인"
		if targetUser, err := u.userRepository.FindByID(ctx, &confirmer.UserID); err == nil && targetUser != nil && targetUser.Name != "" {
			targetUserName = targetUser.Name
		}
		if confirmers, err := u.confirmerRepository.FindAllByUserID(ctx, &confirmer.UserID); err == nil {
			notifiedUsers := make(map[domain.ID]bool)
			for _, conf := range confirmers {
				if conf.Status != domain.UserTransitionConfirmerStatusAccepted {
					continue
				}
				targetUserConf := findUserByEmailOrPhone(ctx, u.userRepository, conf.Email, conf.Phone)
				if targetUserConf != nil && !notifiedUsers[targetUserConf.ID] {
					notifiedUsers[targetUserConf.ID] = true
					_, _ = u.notificationUseCase.Create(ctx, &domain.Notification{
						ReceiverID: targetUserConf.ID,
						Type:       domain.NotificationTypeTransition,
						Message:    fmt.Sprintf("%s님의 사망 확인 심사가 시작되었습니다. 확인을 진행해 주세요.", targetUserName),
						Metadata: map[string]any{
							"type":             "transition_case_started",
							"case_id":          created.ID,
							"target_user_id":   confirmer.UserID,
							"target_user_name": targetUserName,
						},
					})
				}
			}
		}
	}

	return created, nil
}

func (u *userTransitionCaseUseCase) GetByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserTransitionCase, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	return u.evaluateCase(ctx, item, userID)
}

func (u *userTransitionCaseUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.UserTransitionCase, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	items, err := u.repository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	for i, item := range items {
		if item.Status != domain.UserTransitionCaseStatusPending &&
			item.Status != domain.UserTransitionCaseStatusWaiting {
			continue
		}
		evaluated, evaluateErr := u.evaluateCase(ctx, item, userID)
		if evaluateErr != nil {
			return nil, evaluateErr
		}
		items[i] = evaluated
	}
	return items, nil
}

func (u *userTransitionCaseUseCase) ListActiveByUserID(c context.Context, userID *domain.ID) ([]*domain.UserTransitionCase, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindActiveByUserID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return []*domain.UserTransitionCase{}, nil
		}
		return nil, domain.NewInternalServerError(err)
	}

	evaluated, err := u.evaluateCase(ctx, item, userID)
	if err != nil {
		return nil, err
	}
	return []*domain.UserTransitionCase{evaluated}, nil
}

func (u *userTransitionCaseUseCase) Cancel(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserTransitionCase, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	if item.Status == domain.UserTransitionCaseStatusConfirmed {
		return nil, domain.NewBadRequestError(errors.New("confirmed transition case cannot be cancelled"))
	}
	if item.Status == domain.UserTransitionCaseStatusCancelled {
		return item, nil
	}

	now := time.Now()
	item.Status = domain.UserTransitionCaseStatusCancelled
	item.CancelledAt = &now
	updated, err := u.repository.Update(ctx, item, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			current, getErr := u.repository.FindByID(ctx, id, userID)
			if getErr == nil && current.Status == domain.UserTransitionCaseStatusCancelled {
				return current, nil
			}
			return nil, domain.NewBadRequestError(errors.New("confirmed transition case cannot be cancelled"))
		}
		return nil, domain.NewInternalServerError(err)
	}
	return updated, nil
}

func (u *userTransitionCaseUseCase) Evaluate(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserTransitionCase, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	return u.evaluateCase(ctx, item, userID)
}

func (u *userTransitionCaseUseCase) evaluateCase(c context.Context, item *domain.UserTransitionCase, userID *domain.ID) (*domain.UserTransitionCase, error) {
	if item.Status == domain.UserTransitionCaseStatusConfirmed ||
		item.Status == domain.UserTransitionCaseStatusCancelled {
		return item, nil
	}

	now := time.Now()
	if item.Status == domain.UserTransitionCaseStatusPending {
		agreements, err := u.confirmationRepository.CountAgreeByCaseID(c, &item.ID, userID)
		if err != nil {
			return nil, domain.NewInternalServerError(err)
		}
		if agreements < item.MinimumConfirmersSnapshot {
			return item, nil
		}

		eligibleAt := now.Add(time.Duration(item.WaitingPeriodDaysSnapshot) * 24 * time.Hour)
		item.Status = domain.UserTransitionCaseStatusWaiting
		item.WaitingStartedAt = &now
		item.EligibleAt = &eligibleAt
		updated, err := u.repository.Update(c, item, userID)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidTransitionState) {
				current, getErr := u.repository.FindByID(c, &item.ID, userID)
				if getErr == nil {
					return current, nil
				}
			}
			return nil, domain.NewInternalServerError(err)
		}
		item = updated
	}

	if item.Status != domain.UserTransitionCaseStatusWaiting {
		return item, nil
	}
	if item.EligibleAt == nil || now.Before(*item.EligibleAt) {
		return item, nil
	}

	confirmed, err := u.repository.ConfirmAndCreateTasks(c, &item.ID, userID, now)
	if err != nil {
		if errors.Is(err, domain.ErrTransitionCaseNotEligible) {
			current, getErr := u.repository.FindByID(c, &item.ID, userID)
			if getErr == nil &&
				(current.Status == domain.UserTransitionCaseStatusConfirmed ||
					current.Status == domain.UserTransitionCaseStatusCancelled) {
				return current, nil
			}
			return nil, domain.NewBadRequestError(err)
		}
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	if u.notificationUseCase != nil && u.userRepository != nil {
		targetUserName := "지인"
		if targetUser, err := u.userRepository.FindByID(c, userID); err == nil && targetUser != nil && targetUser.Name != "" {
			targetUserName = targetUser.Name
		}
		notifiedUsers := make(map[domain.ID]bool)
		if u.confirmerRepository != nil {
			if confirmers, err := u.confirmerRepository.FindAllByUserID(c, userID); err == nil {
				for _, conf := range confirmers {
					if conf.Status != domain.UserTransitionConfirmerStatusAccepted {
						continue
					}
					targetUserConf := findUserByEmailOrPhone(c, u.userRepository, conf.Email, conf.Phone)
					if targetUserConf != nil && !notifiedUsers[targetUserConf.ID] && targetUserConf.ID != *userID {
						notifiedUsers[targetUserConf.ID] = true
						_, _ = u.notificationUseCase.Create(c, &domain.Notification{
							ReceiverID: targetUserConf.ID,
							Type:       domain.NotificationTypeTransition,
							Message:    fmt.Sprintf("%s님의 사망 전환이 확정되었습니다.", targetUserName),
							Metadata: map[string]any{
								"type":             "transition_confirmed",
								"case_id":          confirmed.ID,
								"target_user_id":   *userID,
								"target_user_name": targetUserName,
							},
						})
					}
				}
			}
		}
		if u.stakeholderRepository != nil {
			if stakeholders, err := u.stakeholderRepository.FindAllByUserID(c, userID); err == nil {
				for _, s := range stakeholders {
					targetUserStakeholder := findUserByEmailOrPhone(c, u.userRepository, s.Email, s.Phone)
					if targetUserStakeholder != nil && !notifiedUsers[targetUserStakeholder.ID] && targetUserStakeholder.ID != *userID {
						notifiedUsers[targetUserStakeholder.ID] = true
						_, _ = u.notificationUseCase.Create(c, &domain.Notification{
							ReceiverID: targetUserStakeholder.ID,
							Type:       domain.NotificationTypeTransition,
							Message:    fmt.Sprintf("%s님의 사망 전환이 확정되었습니다.", targetUserName),
							Metadata: map[string]any{
								"type":             "transition_confirmed",
								"case_id":          confirmed.ID,
								"target_user_id":   *userID,
								"target_user_name": targetUserName,
							},
						})
					}
				}
			}
		}
	}

	return confirmed, nil
}
