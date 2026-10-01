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

func TestFetchByUserID(t *testing.T) {
	mockTaskRepository := new(mocks.TaskRepository)
	userID := domain.NewID()

	t.Run("success", func(t *testing.T) {

		mockTask := domain.Task{
			ID:     domain.NewID(),
			Title:  "Test Title",
			UserID: userID,
		}

		mockListTask := make([]*domain.Task, 0)
		mockListTask = append(mockListTask, &mockTask)

		limit := 10
		offset := 0

		mockPaginatedResult := &domain.PaginatedResult[*domain.Task]{
			Items:  mockListTask,
			Total:  1,
			Limit:  limit,
			Offset: offset,
		}

		mockTaskRepository.On("FetchByUserID", mock.Anything, &userID, offset, limit).Return(mockListTask, 1, nil).Once()

		u := usecase.NewTaskUseCase(mockTaskRepository, time.Second*2)

		result, err := u.ListByUserID(context.Background(), &userID, 0, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, len(mockListTask))
		assert.Equal(t, result, mockPaginatedResult)

		mockTaskRepository.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockTaskRepository.On("FetchByUserID", mock.Anything, &userID, 0, 10).Return(nil, 0, errors.New("unexpected")).Once()

		u := usecase.NewTaskUseCase(mockTaskRepository, time.Second*2)

		list, err := u.ListByUserID(context.Background(), &userID, 0, 10)

		assert.Error(t, err)
		assert.Nil(t, list)

		mockTaskRepository.AssertExpectations(t)
	})

}
