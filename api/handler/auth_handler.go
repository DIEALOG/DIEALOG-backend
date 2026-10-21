package handler

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

// Login godoc
// @Summary      로그인
// @Description  이메일과 비밀번호로 로그인하여 액세스 및 리프레시 토큰을 발급합니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body domain.LoginRequest true "로그인 정보"
// @Success      200 {object} domain.AuthResponse "로그인 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "이메일 또는 비밀번호 불일치"
// @Router       /login [post]
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

// RefreshToken godoc
// @Summary      토큰 갱신
// @Description  리프레시 토큰으로 새로운 액세스 및 리프레시 토큰을 재발급합니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body domain.RefreshTokenRequest true "리프레시 토큰 정보"
// @Success      200 {object} domain.AuthResponse "토큰 재발급 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "유효하지 않은 리프레시 토큰"
// @Router       /refresh [post]
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

// Signup godoc
// @Summary      회원가입
// @Description  새로운 사용자 계정을 생성하고 액세스 및 리프레시 토큰을 발급합니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body domain.SignupRequest true "회원가입 정보"
// @Success      200 {object} domain.AuthResponse "회원가입 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /signup [post]
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
