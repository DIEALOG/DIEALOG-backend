package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type userTransitionConfirmationUseCase struct {
	repository     domain.UserTransitionConfirmationRepository
	caseRepository domain.UserTransitionCaseRepository
	caseUseCase    domain.UserTransitionCaseUseCase
	contextTimeout time.Duration
}

func NewUserTransitionConfirmationUseCase(repository domain.UserTransitionConfirmationRepository, caseRepository domain.UserTransitionCaseRepository, caseUseCase domain.UserTransitionCaseUseCase, timeout time.Duration) domain.UserTransitionConfirmationUseCase {
	return &userTransitionConfirmationUseCase{
		repository:     repository,
		caseRepository: caseRepository,
		caseUseCase:    caseUseCase,
		contextTimeout: timeout,
	}
}

func (u *userTransitionConfirmationUseCase) Submit(c context.Context, caseID *domain.ID, confirmer *domain.UserTransitionConfirmer, request *domain.SubmitTransitionConfirmationRequest) (*domain.UserTransitionConfirmation, error) {
	if confirmer == nil {
		return nil, domain.NewUnauthorizedError(errors.New("confirmer is required"))
	}
	if confirmer.Status != domain.UserTransitionConfirmerStatusAccepted {
		return nil, domain.NewForbiddenError(domain.ErrConfirmerNotAccepted)
	}
	if request == nil {
		return nil, domain.NewBadRequestError(errors.New("request is required"))
	}
	switch request.Decision {
	case domain.UserTransitionDecisionAgree, domain.UserTransitionDecisionDisagree, domain.UserTransitionDecisionAbstain:
	default:
		return nil, domain.NewBadRequestError(errors.New("decision must be agree, disagree, or abstain"))
	}

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	transitionCase, err := u.caseRepository.FindByID(ctx, caseID, &confirmer.UserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	if transitionCase.Status != domain.UserTransitionCaseStatusPending {
		return nil, domain.NewBadRequestError(errors.New("confirmations can only be submitted while the case is pending"))
	}

	now := time.Now()
	saved, err := u.repository.Upsert(ctx, &domain.UserTransitionConfirmation{
		TransitionCaseID:      *caseID,
		TransitionConfirmerID: confirmer.ID,
		Decision:              request.Decision,
		Comment:               request.Comment,
		RespondedAt:           &now,
	}, &confirmer.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrTransitionCaseNotEligible) {
			return nil, domain.NewBadRequestError(errors.New("confirmations can only be submitted while the case is pending"))
		}
		if errors.Is(err, domain.ErrConfirmerNotAccepted) {
			return nil, domain.NewForbiddenError(domain.ErrConfirmerNotAccepted)
		}
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	if _, err := u.caseUseCase.Evaluate(ctx, caseID, &confirmer.UserID); err != nil {
		return nil, err
	}

	return saved, nil
}

func (u *userTransitionConfirmationUseCase) ListByCaseID(c context.Context, caseID *domain.ID, userID *domain.ID) ([]*domain.UserTransitionConfirmation, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	if _, err := u.caseRepository.FindByID(ctx, caseID, userID); err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	items, err := u.repository.FindByCaseID(ctx, caseID, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	return items, nil
}
