package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/handler"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/middleware"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func NewProfileRouter(app fiber.Router, service domain.ProfileUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)
	protected.Get("/", handler.FetchProfile(service))
}
