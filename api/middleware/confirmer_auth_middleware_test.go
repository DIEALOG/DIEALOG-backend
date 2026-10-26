package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/domain/mocks"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConfirmerMiddleware(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		service := new(mocks.UserTransitionConfirmerUseCase)
		app := fiber.New()
		app.Get("/", middleware.ConfirmerMiddleware(service), func(c fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})

		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		service.AssertNotCalled(t, "Authenticate", mock.Anything, mock.Anything)
	})

	t.Run("invalid token", func(t *testing.T) {
		service := new(mocks.UserTransitionConfirmerUseCase)
		service.On("Authenticate", mock.Anything, "invalid").
			Return(nil, domain.NewUnauthorizedError(errors.New("invalid confirmer token"))).
			Once()

		app := fiber.New()
		app.Get("/", middleware.ConfirmerMiddleware(service), func(c fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Confirmer-Token", "invalid")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		confirmer := &domain.UserTransitionConfirmer{
			ID:     domain.NewID(),
			UserID: domain.NewID(),
			Status: domain.UserTransitionConfirmerStatusAccepted,
		}
		service := new(mocks.UserTransitionConfirmerUseCase)
		service.On("Authenticate", mock.Anything, "valid").
			Return(confirmer, nil).
			Once()

		app := fiber.New()
		app.Get("/", middleware.ConfirmerMiddleware(service), func(c fiber.Ctx) error {
			got := c.Locals("confirmer").(*domain.UserTransitionConfirmer)
			return c.JSON(fiber.Map{"id": got.ID})
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Confirmer-Token", "valid")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		service.AssertExpectations(t)
	})
}
