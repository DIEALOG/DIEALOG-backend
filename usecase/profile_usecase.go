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
		Name:              user.Name,
		Email:             user.Email,
		PhoneNumber:       user.PhoneNumber,
		Stakeholders:      user.Stakeholders,
		StakeholderGroups: user.StakeholderGroups,
	}, nil
}
