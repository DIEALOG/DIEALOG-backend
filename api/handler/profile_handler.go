package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func FetchProfile(service domain.ProfileUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		profile, err := service.GetProfileByID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(profile)
	}
}

func UpdateTestament(service domain.ProfileUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		var req domain.UpdateTestamentRequest
		err := c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		err = service.UpdateTestament(ctx, userID, req.Testament)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Testament updated successfully",
		})
	}
}
