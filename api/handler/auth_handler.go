package handler

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func Login(service domain.AuthUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var request domain.LoginRequest

		err := c.Bind().Body(&request)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		user, err := service.Login(ctx, request.Email, request.Password)
		if err != nil {
			if errInfo, ok := errors.AsType[domain.Error](err); ok {
				return c.Status(errInfo.StatusCode).JSON(domain.ErrorResponse{Message: err.Error()})
			}
		}

		accessToken, refreshToken, err := service.CreateAccessAndRefreshToken(ctx, user)
		if err != nil {
			if errInfo, ok := errors.AsType[domain.Error](err); ok {
				return c.Status(errInfo.StatusCode).JSON(domain.ErrorResponse{Message: err.Error()})
			}
		}

		response := domain.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		return c.Status(http.StatusOK).JSON(response)
	}
}

func RefreshToken(service domain.AuthUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var request domain.RefreshTokenRequest

		err := c.Bind().Body(&request)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		user, err := service.ExtractUserFromRefreshToken(ctx, request.RefreshToken)
		if err != nil {
			if errInfo, ok := errors.AsType[domain.Error](err); ok {
				return c.Status(errInfo.StatusCode).JSON(domain.ErrorResponse{Message: err.Error()})
			}
		}

		accessToken, refreshToken, err := service.CreateAccessAndRefreshToken(ctx, user)
		if err != nil {
			if errInfo, ok := errors.AsType[domain.Error](err); ok {
				return c.Status(errInfo.StatusCode).JSON(domain.ErrorResponse{Message: err.Error()})
			}
		}

		response := domain.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		return c.Status(http.StatusOK).JSON(response)
	}
}

func Signup(service domain.AuthUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var request domain.SignupRequest

		err := c.Bind().Body(&request)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		user, err := service.Register(ctx, request.Name, request.Email, request.Password)
		if err != nil {
			return RespondError(c, err)
		}

		accessToken, refreshToken, err := service.CreateAccessAndRefreshToken(ctx, user)
		if err != nil {
			return RespondError(c, err)
		}

		response := domain.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		return c.Status(http.StatusOK).JSON(response)
	}
}
