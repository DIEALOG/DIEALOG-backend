package usecase

import (
	"context"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type profileUseCase struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func NewProfileUseCase(userRepository domain.UserRepository, timeout time.Duration) domain.ProfileUseCase {
	return &profileUseCase{
		userRepository: userRepository,
		contextTimeout: timeout,
	}
}

func (pu *profileUseCase) GetProfileByID(c context.Context, userID *domain.ID) (*domain.Profile, error) {
	ctx, cancel := context.WithTimeout(c, pu.contextTimeout)
	defer cancel()

	user, err := pu.userRepository.FindByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.Profile{
		Name:         user.Name,
		Email:        user.Email,
		Testament:    user.Testament,
		Stakeholders: user.Stakeholders,
	}, nil
}

func (pu *profileUseCase) UpdateTestament(c context.Context, userID *domain.ID, testament string) error {
	ctx, cancel := context.WithTimeout(c, pu.contextTimeout)
	defer cancel()

	err := pu.userRepository.UpdateTestament(ctx, userID, testament)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}
