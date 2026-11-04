package usecase

import (
	"context"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
)

type userUseCase struct {
	userRepository domain.UserRepository
	timeout        time.Duration
}

func NewUserUseCase(userRepository domain.UserRepository, timeout time.Duration) domain.UserUseCase {
	return &userUseCase{
		userRepository: userRepository,
		timeout:        timeout,
	}
}

func (uu *userUseCase) Update(c context.Context, user *domain.User) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(c, uu.timeout)
	defer cancel()

	return uu.userRepository.Update(ctx, user)
}

func (uu *userUseCase) Delete(c context.Context, id *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, uu.timeout)
	defer cancel()

	return uu.userRepository.Delete(ctx, id)
}