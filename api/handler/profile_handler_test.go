package handler_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/handler"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain/mocks"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setUserID(userID domain.ID) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("id", &userID)
		return c.Next()
	}
}

func TestFetch(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		mockProfile := &domain.Profile{
			Name:  "Test Name",
			Email: "test@gmail.com",
		}

		userID := domain.NewID()

		mockProfileUseCase := new(mocks.ProfileUseCase)

		mockProfileUseCase.On("GetProfileByID", mock.Anything, &userID).Return(mockProfile, nil)

		app := fiber.New()

		app.Use(setUserID(userID))
		app.Get("/profile", handler.FetchProfile(mockProfileUseCase))

		body, err := json.Marshal(mockProfile)
		assert.NoError(t, err)

		bodyString := string(body)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, bodyString, string(bodyBytes))

		mockProfileUseCase.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		userID := domain.NewID()

		mockProfileUseCase := new(mocks.ProfileUseCase)

		customErr := errors.New("unexpected")

		mockProfileUseCase.On("GetProfileByID", mock.Anything, &userID).Return(nil, customErr)

		app := fiber.New()

		app.Use(setUserID(userID))
		app.Get("/profile", handler.FetchProfile(mockProfileUseCase))

		body, err := json.Marshal(domain.ErrorResponse{Message: customErr.Error()})
		assert.NoError(t, err)

		bodyString := string(body)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, bodyString, string(bodyBytes))

		mockProfileUseCase.AssertExpectations(t)
	})

}
