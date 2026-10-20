package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/handler"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/middleware"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

func NewStakeholderRouter(app fiber.Router, stakeholderService domain.StakeholderUseCase, groupService domain.StakeholderGroupUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)

	// stakeholder group routes under /group
	group := protected.Group("group")
	group.Post("/", handler.CreateStakeholderGroup(groupService))
	group.Get("/", handler.FetchStakeholderGroups(groupService))
	group.Get("/:id", handler.GetStakeholderGroup(groupService))
	group.Put("/:id", handler.UpdateStakeholderGroup(groupService))
	group.Delete("/:id", handler.DeleteStakeholderGroup(groupService))

	// stakeholder routes
	protected.Post("/", handler.CreateStakeholder(stakeholderService))
	protected.Get("/", handler.FetchStakeholders(stakeholderService))
	protected.Get("/:id", handler.GetStakeholder(stakeholderService))
	protected.Put("/:id", handler.UpdateStakeholder(stakeholderService))
	protected.Delete("/:id", handler.DeleteStakeholder(stakeholderService))
}
