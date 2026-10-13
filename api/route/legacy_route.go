package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/handler"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/middleware"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func NewLegacyRouter(app fiber.Router, service domain.LegacyUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)

	protected.Post("/", handler.CreateLegacy(service))
	protected.Get("/:id", handler.GetLegacy(service))
	protected.Put("/:id", handler.UpdateLegacy(service))
	protected.Delete("/:id", handler.DeleteLegacy(service))

	// account endpoints under legacy
	protected.Post("/:id/account", handler.CreateAccount(service))
	protected.Delete("/:id/account/:accountId", handler.DeleteAccount(service))

	// paginated list
	protected.Use(paginate.New())
	protected.Get("/", handler.FetchLegacies(service))
}
