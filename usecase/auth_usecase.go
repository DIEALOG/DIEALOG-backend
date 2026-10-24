package usecase

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/config"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepository       domain.UserRepository
	userConfigRepository domain.UserConfigRepository
	contextTimeout       time.Duration
}

func NewAuthUseCase(userRepository domain.UserRepository, userConfigRepository domain.UserConfigRepository, timeout time.Duration) domain.AuthUseCase {
	return &authUseCase{
		userRepository:       userRepository,
		userConfigRepository: userConfigRepository,
		contextTimeout:       timeout,
	}
}

func (u *authUseCase) Register(c context.Context, name, email, password, phoneNumber string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	var phoneRegex = regexp.MustCompile(`^01[016789]-?\d{3,4}-?\d{4}$`)

	if name == "" {
		return nil, domain.NewBadRequestError(errors.New("name is required"))
	}
	if email == "" {
		return nil, domain.NewBadRequestError(errors.New("email is required"))
	}
	if phoneNumber == "" || !phoneRegex.MatchString(phoneNumber) {
		return nil, domain.NewBadRequestError(errors.New("invalid phone number"))
	}
	if password == "" {
		return nil, domain.NewBadRequestError(errors.New("password is required"))
	}

	_, err := u.userRepository.FindByEmail(ctx, email)
	if err == nil {
		return nil, domain.NewBadRequestError(err)
	}

	encrypted, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, domain.NewBadRequestError(err)
	}

	user, err := u.userRepository.Create(ctx, &domain.User{
		Name:        name,
		Email:       email,
		Password:    string(encrypted),
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	_, err = u.userConfigRepository.Create(ctx, &domain.UserConfig{
		UserID: user.ID,
	})
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return user, nil
}

func (u *authUseCase) Login(c context.Context, email, password string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	user, err := u.userRepository.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.NewUnauthorizedError(err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, domain.NewBadRequestError(errors.New("invalid credentials"))
	}

	return user, nil
}

func (u *authUseCase) CreateAccessAndRefreshToken(c context.Context, user *domain.User) (string, string, error) {
	access, err := token.CreateAccessToken(user, config.E.AccessTokenSecret, config.E.AccessTokenExpiryHour)
	if err != nil {
		return "", "", domain.NewInternalServerError(err)
	}

	refresh, err := token.CreateRefreshToken(user, config.E.RefreshTokenSecret, config.E.RefreshTokenExpiryHour)
	if err != nil {
		return "", "", domain.NewInternalServerError(err)
	}

	return access, refresh, nil
}

func (u *authUseCase) ExtractUserFromRefreshToken(c context.Context, requestToken string) (*domain.User, error) {
	id, err := token.ExtractIDFromToken(requestToken, config.E.RefreshTokenSecret)
	if err != nil {
		return nil, domain.NewBadRequestError(err)
	}

	user, err := u.userRepository.FindByID(c, id)
	if err != nil {
		return nil, domain.NewUnauthorizedError(err)
	}

	return user, nil
}
