package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/domain/mocks"
	"github.com/DIEALOG/DIEALOG-backend/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchLegacyByUserID(t *testing.T) {
	mockLegacyRepository := new(mocks.LegacyRepository)
	mockCredentialUseCase := new(mocks.CredentialUseCase)
	userID := domain.NewID()

	t.Run("success", func(t *testing.T) {
		mockLegacy := domain.Legacy{
			ID:     domain.NewID(),
			Name:   "Google Account",
			UserID: userID,
		}

		mockListLegacy := []*domain.Legacy{&mockLegacy}

		mockLegacyRepository.
			On("FindAllByUserID", mock.Anything, &userID).
			Return(mockListLegacy, nil).
			Once()

		u := usecase.NewLegacyUseCase(mockLegacyRepository, mockCredentialUseCase, nil, time.Second*2)

		result, err := u.ListByUserID(context.Background(), &userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, len(mockListLegacy))
		assert.Equal(t, mockListLegacy, result)

		mockLegacyRepository.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockLegacyRepository.
			On("FindAllByUserID", mock.Anything, &userID).
			Return(nil, errors.New("unexpected")).
			Once()

		u := usecase.NewLegacyUseCase(mockLegacyRepository, mockCredentialUseCase, nil, time.Second*2)

		list, err := u.ListByUserID(context.Background(), &userID)

		assert.Error(t, err)
		assert.Nil(t, list)

		mockLegacyRepository.AssertExpectations(t)
	})
}
