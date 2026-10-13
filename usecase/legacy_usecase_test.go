package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain/mocks"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchLegacyByUserID(t *testing.T) {
	mockLegacyRepository := new(mocks.LegacyRepository)
	userID := domain.NewID()

	t.Run("success", func(t *testing.T) {
		mockLegacy := domain.Legacy{
			ID:     domain.NewID(),
			Name:   "Google Account",
			UserID: userID,
		}

		mockListLegacy := make([]*domain.Legacy, 0)
		mockListLegacy = append(mockListLegacy, &mockLegacy)

		limit := 10
		offset := 0

		mockPaginatedResult := &domain.PaginatedResult[*domain.Legacy]{
			Items:  mockListLegacy,
			Total:  1,
			Limit:  limit,
			Offset: offset,
		}

		mockLegacyRepository.On("FindAllByUserID", mock.Anything, &userID, offset, limit).Return(mockListLegacy, 1, nil).Once()

		u := usecase.NewLegacyUseCase(mockLegacyRepository, time.Second*2)

		result, err := u.ListByUserID(context.Background(), &userID, 0, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, len(mockListLegacy))
		assert.Equal(t, mockPaginatedResult, result)

		mockLegacyRepository.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockLegacyRepository.On("FindAllByUserID", mock.Anything, &userID, 0, 10).Return(nil, 0, errors.New("unexpected")).Once()

		u := usecase.NewLegacyUseCase(mockLegacyRepository, time.Second*2)

		list, err := u.ListByUserID(context.Background(), &userID, 0, 10)

		assert.Error(t, err)
		assert.Nil(t, list)

		mockLegacyRepository.AssertExpectations(t)
	})
}
