package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewUserRouter(app fiber.Router, usecase domain.UserUseCase) {
	// protected
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)
	protected.Put("/", handler.UpdateUser(usecase))
	protected.Delete("/", handler.DeleteUser(usecase))
}