package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewLastWordRouter(app fiber.Router, service domain.LastWordUseCase, stakeholderService domain.StakeholderUseCase, stakeholderGroupService domain.StakeholderGroupUseCase) {
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)

	protected.Post("/", handler.CreateLastWord(service, stakeholderService, stakeholderGroupService))
	protected.Get("/", handler.FetchMyLastWords(service))
	protected.Get("/:id", handler.GetLastWord(service))
	protected.Put("/:id", handler.UpdateLastWord(service, stakeholderService, stakeholderGroupService))
	protected.Delete("/:id", handler.DeleteLastWord(service))
}
