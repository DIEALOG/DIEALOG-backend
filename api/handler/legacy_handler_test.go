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

func TestUpdateAccountPassesPassword(t *testing.T) {
	userID := domain.NewID()
	legacyID := domain.NewID()
	accountID := domain.NewID()
	password := "new-password"

	service := new(mocks.LegacyUseCase)
	stakeholderService := new(mocks.StakeholderUseCase)
	stakeholderGroupService := new(mocks.StakeholderGroupUseCase)

	stakeholderService.On("ListByIDs", mock.Anything, mock.Anything, mock.Anything).
		Return([]*domain.Stakeholder{}, nil).
		Once()
	stakeholderGroupService.On("ListByIDs", mock.Anything, mock.Anything, mock.Anything).
		Return([]*domain.StakeholderGroup{}, nil).
		Once()
	service.On("UpdateAccount",
		mock.Anything,
		mock.MatchedBy(func(acc *domain.Account) bool {
			return acc != nil && acc.ID == accountID && acc.Name == "Kakao"
		}),
		mock.MatchedBy(func(id *domain.ID) bool { return id != nil && *id == legacyID }),
		mock.MatchedBy(func(id *domain.ID) bool { return id != nil && *id == userID }),
		mock.MatchedBy(func(p *string) bool { return p != nil && *p == password }),
	).
		Return(&domain.Account{
			ID:       accountID,
			LegacyID: legacyID,
			Name:     "Kakao",
			Type:     domain.AccountType{Type: domain.AccountTypeKakaoTalk},
			Action:   domain.AccountAction{Action: domain.AccountActionRemain},
		}, nil).
		Once()

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("id", &userID)
		return c.Next()
	})
	app.Put("/legacy/protected/:id/account/:accountId", handler.UpdateAccount(service, stakeholderService, stakeholderGroupService))

	body := `{"type":{"type":"KakaoTalk"},"name":"Kakao","action":{"action":"remain"},"account_password":"` + password + `"}`
	req := httptest.NewRequest(
		http.MethodPut,
		"/legacy/protected/"+legacyID.String()+"/account/"+accountID.String(),
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(respBody), accountID.String())
	assert.Contains(t, string(respBody), password)

	service.AssertExpectations(t)
	stakeholderService.AssertExpectations(t)
	stakeholderGroupService.AssertExpectations(t)
}
