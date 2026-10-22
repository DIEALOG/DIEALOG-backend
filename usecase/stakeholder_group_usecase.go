package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type stakeholderGroupUseCase struct {
	groupRepository domain.StakeholderGroupRepository
	contextTimeout  time.Duration
}

func NewStakeholderGroupUseCase(groupRepository domain.StakeholderGroupRepository, timeout time.Duration) domain.StakeholderGroupUseCase {
	return &stakeholderGroupUseCase{
		groupRepository: groupRepository,
		contextTimeout:  timeout,
	}
}

func (gu *stakeholderGroupUseCase) Create(c context.Context, g *domain.StakeholderGroup, userID *domain.ID) (*domain.StakeholderGroup, error) {
	if g.Name == "" {
		return nil, domain.NewBadRequestError(errors.New("group name is required"))
	}

	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	created, err := gu.groupRepository.Create(ctx, g, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return created, nil
}

func (gu *stakeholderGroupUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.StakeholderGroup, error) {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	list, err := gu.groupRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return list, nil
}

func (gu *stakeholderGroupUseCase) GetByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.StakeholderGroup, error) {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	item, err := gu.groupRepository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return item, nil
}

func (gu *stakeholderGroupUseCase) Update(c context.Context, id *domain.ID, g *domain.StakeholderGroup, userID *domain.ID) (*domain.StakeholderGroup, error) {
	if g.Name == "" {
		return nil, domain.NewBadRequestError(errors.New("group name is required"))
	}

	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	updated, err := gu.groupRepository.Update(ctx, id, g, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return updated, nil
}

func (gu *stakeholderGroupUseCase) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	err := gu.groupRepository.Delete(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}
