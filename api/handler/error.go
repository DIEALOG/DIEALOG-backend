package handler

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func RespondError(c fiber.Ctx, err error) error {
	status := http.StatusInternalServerError

	if errInfo, ok := errors.AsType[*domain.Error](err); ok {
		status = errInfo.StatusCode
	}

	return c.Status(status).JSON(domain.ErrorResponse{
		Message: err.Error(),
	})
}
