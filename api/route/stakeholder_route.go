package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/handler"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/middleware"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func NewStakeholderRouter(app fiber.Router, service domain.StakeholderUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)
	protected.Post("/", handler.CreateStakeholder(service))
	protected.Get("/", handler.FetchStakeholders(service))
	protected.Delete("/:id", handler.DeleteStakeholder(service))
}
