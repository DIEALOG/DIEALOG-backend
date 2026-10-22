package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewSignupRouter(app fiber.Router, service domain.AuthUseCase) {
	app.Post("/", handler.Signup(service))
}
