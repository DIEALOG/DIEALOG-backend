package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

type UpdateLegacyRequest struct {
	Name string `form:"name" binding:"required" json:"name"`
}

func CreateLegacy(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var legacy domain.Legacy
		err := c.Bind().Body(&legacy)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		created, err := service.Create(ctx, &legacy, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(created)
	}
}

func FetchLegacies(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		pageInfo, ok := paginate.FromContext(c)
		if !ok {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "Pagination failed"})
		}

		legacies, err := service.ListByUserID(ctx, userID, pageInfo.Start(), pageInfo.Limit)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(legacies)
	}
}

func GetLegacy(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		legacy, err := service.GetByID(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(legacy)
	}
}

func UpdateLegacy(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		var req UpdateLegacyRequest
		err = c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		updated, err := service.Update(ctx, &id, req.Name, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(updated)
	}
}

func DeleteLegacy(service domain.LegacyUseCase) fiber.Handler {
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
			Message: "Legacy deleted successfully",
		})
	}
}

func CreateAccount(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		legacyIDStr := c.Params("id")
		legacyID, err := domain.StringToID(legacyIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid legacy id format"})
		}

		var acc domain.Account
		err = c.Bind().Body(&acc)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		created, err := service.CreateAccount(ctx, &acc, &legacyID, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(created)
	}
}

func DeleteAccount(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		legacyIDStr := c.Params("id")
		legacyID, err := domain.StringToID(legacyIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid legacy id format"})
		}

		accountIDStr := c.Params("accountId")
		accountID, err := domain.StringToID(accountIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid account id format"})
		}

		err = service.DeleteAccount(ctx, &accountID, &legacyID, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Account deleted successfully",
		})
	}
}
