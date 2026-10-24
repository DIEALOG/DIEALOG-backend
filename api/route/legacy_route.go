package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewLegacyRouter(app fiber.Router, service domain.LegacyUseCase, stakeholderService domain.StakeholderUseCase, stakeholderGroupService domain.StakeholderGroupUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)

	protected.Post("/", handler.CreateLegacy(service))
	protected.Get("/", handler.FetchLegacies(service))
	protected.Get("/:id", handler.GetLegacy(service))
	protected.Put("/:id", handler.UpdateLegacy(service))
	protected.Delete("/:id", handler.DeleteLegacy(service))

	// account endpoints under legacy
	protected.Post("/:id/account", handler.CreateAccount(service, stakeholderService, stakeholderGroupService))
	protected.Get("/:id/account/:accountId", handler.GetAccount(service))
	protected.Delete("/:id/account/:accountId", handler.DeleteAccount(service))
}
