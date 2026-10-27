package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewPhoneVerificationRoute(app fiber.Router, service domain.PhoneVerificationUseCase) {
	app.Post("/send", handler.SendPhoneVerification(service))
	app.Post("/verify", handler.VerifyPhoneVerification(service))
}
