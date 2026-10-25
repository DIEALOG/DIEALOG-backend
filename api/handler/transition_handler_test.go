package handler_test

import (
	"io"
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

func TestAcceptTransitionInvitation(t *testing.T) {
	confirmerID := domain.NewID()
	service := new(mocks.UserTransitionConfirmerUseCase)
	service.On("Accept", mock.Anything, "invitation-token").
		Return(&domain.AcceptTransitionInvitationResponse{
			Confirmer: &domain.UserTransitionConfirmer{
				ID:     confirmerID,
				Status: domain.UserTransitionConfirmerStatusAccepted,
			},
			AuthToken: "auth-token",
		}, nil).
		Once()

	app := fiber.New()
	app.Post("/invitations/accept", handler.AcceptTransitionInvitation(service))

	req := httptest.NewRequest(
		http.MethodPost,
		"/invitations/accept",
		strings.NewReader(`{"token":"invitation-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "auth-token")
	assert.Contains(t, string(body), confirmerID.String())
	service.AssertExpectations(t)
}

func TestFetchTransitionCaseRejectsInvalidUUID(t *testing.T) {
	service := new(mocks.UserTransitionCaseUseCase)
	app := fiber.New()
	app.Get("/cases/:id", handler.FetchTransitionCase(service))

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/cases/not-a-uuid", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	service.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitAccountHandlingResult(t *testing.T) {
	taskID := domain.NewID()
	service := new(mocks.AccountHandlingUseCase)
	service.On("SubmitResult", mock.Anything, mock.MatchedBy(func(request *domain.SubmitAccountHandlingResultRequest) bool {
		return request.Token == "request-token" && request.Success != nil && *request.Success
	})).
		Return(&domain.AccountHandlingTask{
			ID:     taskID,
			Status: domain.AccountHandlingTaskStatusCompleted,
			Assignees: []domain.AccountHandlingTaskAssignee{{
				Name:  "Other Stakeholder",
				Email: stringPointer("other@example.com"),
			}},
			Events: []domain.AccountHandlingEvent{{
				Message: stringPointer("internal history"),
			}},
		}, nil).
		Once()

	app := fiber.New()
	app.Post("/stakeholder-tasks/result", handler.SubmitAccountHandlingResult(service))

	req := httptest.NewRequest(
		http.MethodPost,
		"/stakeholder-tasks/result",
		strings.NewReader(`{"token":"request-token","success":true}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), taskID.String())
	assert.Contains(t, string(body), string(domain.AccountHandlingTaskStatusCompleted))
	assert.NotContains(t, string(body), "other@example.com")
	assert.NotContains(t, string(body), "internal history")
	service.AssertExpectations(t)
}

func stringPointer(value string) *string {
	return &value
}
