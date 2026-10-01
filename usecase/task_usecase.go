package usecase

import (
	"context"
	"time"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

type taskUseCase struct {
	taskRepository domain.TaskRepository
	contextTimeout time.Duration
}

func NewTaskUseCase(taskRepository domain.TaskRepository, timeout time.Duration) domain.TaskUseCase {
	return &taskUseCase{
		taskRepository: taskRepository,
		contextTimeout: timeout,
	}
}

func (tu *taskUseCase) Create(c context.Context, task *domain.Task, userID *domain.ID) (*domain.Task, error) {
	ctx, cancel := context.WithTimeout(c, tu.contextTimeout)
	defer cancel()
	task.UserID = *userID

	t, err := tu.taskRepository.Create(ctx, task)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return t, nil
}

func (tu *taskUseCase) ListByUserID(c context.Context, userID *domain.ID, offset, limit int) (*domain.PaginatedResult[*domain.Task], error) {
	ctx, cancel := context.WithTimeout(c, tu.contextTimeout)
	defer cancel()
	tasks, total, err := tu.taskRepository.FindAllByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.PaginatedResult[*domain.Task]{
		Items:  tasks,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
