package middleware

import (
	"errors"
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func ConfirmerMiddleware(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Get("X-Confirmer-Token")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{
				Message: "X-Confirmer-Token header is empty",
			})
		}

		confirmer, err := service.Authenticate(c.RequestCtx(), token)
		if err != nil {
			status := http.StatusUnauthorized
			if errInfo, ok := errors.AsType[*domain.Error](err); ok {
				status = errInfo.StatusCode
			}
			return c.Status(status).JSON(domain.ErrorResponse{Message: err.Error()})
		}
		if confirmer == nil {
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{Message: "invalid confirmer token"})
		}

		c.Locals("confirmer", confirmer)
		c.Locals("confirmer_id", &confirmer.ID)
		return c.Next()
	}
}
