package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func CreateStakeholder(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var stakeholder domain.Stakeholder
		err := c.Bind().Body(&stakeholder)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		created, err := service.Create(ctx, &stakeholder, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(created)
	}
}

func FetchStakeholders(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		list, err := service.ListByUserID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(list)
	}
}

func DeleteStakeholder(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		err = service.Delete(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Stakeholder deleted successfully",
		})
	}
}
