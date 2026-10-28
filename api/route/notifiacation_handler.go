package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewNotificationRoute(app fiber.Router, service domain.NotificationUseCase) {
	protected := app.Group("/protected")
	protected.Use(middleware.JwtMiddleware)
	protected.Get("/", handler.GetNotifications(service))
	protected.Patch("/:id", handler.MarkNotificationAsRead(service))
	protected.Delete("/:id", handler.DeleteNotification(service))
	protected.Get("/unread-count", handler.GetUnreadNotificationsCount(service))
}
