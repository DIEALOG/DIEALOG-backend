package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewProfileRouter(app fiber.Router, service domain.ProfileUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)
	protected.Get("/", handler.FetchProfile(service))
	protected.Put("/testament", handler.UpdateTestament(service))
}
