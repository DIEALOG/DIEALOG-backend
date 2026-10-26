package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type stakeholderUseCase struct {
	stakeholderRepository domain.StakeholderRepository
	contextTimeout        time.Duration
}

func NewStakeholderUseCase(stakeholderRepository domain.StakeholderRepository, timeout time.Duration) domain.StakeholderUseCase {
	return &stakeholderUseCase{
		stakeholderRepository: stakeholderRepository,
		contextTimeout:        timeout,
	}
}

func (su *stakeholderUseCase) Create(c context.Context, s *domain.Stakeholder, userID *domain.ID) (*domain.Stakeholder, error) {
	if !s.IsValid() {
		return nil, domain.NewBadRequestError(errors.New("name is required; either phone or email must be provided"))
	}

	ctx, cancel := context.WithTimeout(c, su.contextTimeout)
	defer cancel()

	created, err := su.stakeholderRepository.Create(ctx, s, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return created, nil
}

func (su *stakeholderUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.Stakeholder, error) {
	ctx, cancel := context.WithTimeout(c, su.contextTimeout)
	defer cancel()

	list, err := su.stakeholderRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return list, nil
}

func (su *stakeholderUseCase) GetByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.Stakeholder, error) {
	ctx, cancel := context.WithTimeout(c, su.contextTimeout)
	defer cancel()

	item, err := su.stakeholderRepository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return item, nil
}

func (su *stakeholderUseCase) Update(c context.Context, id *domain.ID, s *domain.Stakeholder, userID *domain.ID) (*domain.Stakeholder, error) {
	if !s.IsValid() {
		return nil, domain.NewBadRequestError(errors.New("name is required; either phone or email must be provided"))
	}

	ctx, cancel := context.WithTimeout(c, su.contextTimeout)
	defer cancel()

	updated, err := su.stakeholderRepository.Update(ctx, id, s, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return updated, nil
}

func (su *stakeholderUseCase) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, su.contextTimeout)
	defer cancel()

	err := su.stakeholderRepository.Delete(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}

func (su *stakeholderUseCase) ListByIDs(c context.Context, ids []domain.ID, userID *domain.ID) ([]*domain.Stakeholder, error) {
	ctx, cancel := context.WithTimeout(c, su.contextTimeout)
	defer cancel()

	stakeholders, err := su.stakeholderRepository.FindAllByIDs(ctx, ids, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return stakeholders, nil
}
