package usecase

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
)

func findUserByEmailOrPhone(ctx context.Context, userRepo domain.UserRepository, email *string, phone *string) *domain.User {
	if userRepo == nil {
		return nil
	}
	if email != nil && *email != "" {
		if u, err := userRepo.FindByEmail(ctx, *email); err == nil && u != nil {
			return u
		}
	}
	if phone != nil && *phone != "" {
		if u, err := userRepo.FindByPhoneNumber(ctx, *phone); err == nil && u != nil {
			return u
		}
	}
	return nil
}
