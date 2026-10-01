package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/handler"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func NewRefreshTokenRouter(app fiber.Router, service domain.AuthUseCase) {
	app.Post("/", handler.RefreshToken(service))
}
