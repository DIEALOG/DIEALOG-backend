package handler

import (
	"errors"
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
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
