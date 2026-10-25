package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/domain/mocks"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoginReturnsDomainError(t *testing.T) {
	mockAuthUseCase := new(mocks.AuthUseCase)
	mockAuthUseCase.
		On("Login", mock.Anything, "test@example.com", "wrong").
		Return(nil, domain.NewUnauthorizedError(errors.New("invalid credentials"))).
		Once()

	app := fiber.New()
	app.Post("/login", handler.Login(mockAuthUseCase))

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"email":"test@example.com","password":"wrong"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	mockAuthUseCase.AssertNotCalled(t, "CreateAccessAndRefreshToken", mock.Anything, mock.Anything)
	mockAuthUseCase.AssertExpectations(t)
}

func TestRefreshTokenReturnsDomainError(t *testing.T) {
	mockAuthUseCase := new(mocks.AuthUseCase)
	mockAuthUseCase.
		On("ExtractUserFromRefreshToken", mock.Anything, "invalid-token").
		Return(nil, domain.NewBadRequestError(errors.New("invalid refresh token"))).
		Once()

	app := fiber.New()
	app.Post("/refresh", handler.RefreshToken(mockAuthUseCase))

	req := httptest.NewRequest(
		http.MethodPost,
		"/refresh",
		strings.NewReader(`{"refreshToken":"invalid-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockAuthUseCase.AssertNotCalled(t, "CreateAccessAndRefreshToken", mock.Anything, mock.Anything)
	mockAuthUseCase.AssertExpectations(t)
}
