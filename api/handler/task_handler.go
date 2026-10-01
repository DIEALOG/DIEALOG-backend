package handler

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func CreateTask(service domain.TaskUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var task domain.Task

		err := c.Bind().Body(&task)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		_, err = service.Create(ctx, &task, userID)
		if err != nil {
			if errInfo, ok := errors.AsType[domain.Error](err); ok {
				return c.Status(errInfo.StatusCode).JSON(domain.ErrorResponse{Message: err.Error()})
			}
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Task created successfully",
		})
	}
}

func FetchTask(service domain.TaskUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		userID := c.Locals("id").(*domain.ID)

		pageInfo, ok := paginate.FromContext(c)
		if !ok {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "Pagination failed"})
		}

		tasks, err := service.ListByUserID(ctx, userID, pageInfo.Start(), pageInfo.Limit)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(tasks)
	}
}
