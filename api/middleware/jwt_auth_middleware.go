package middleware

import (
	"net/http"
	"strings"

	"github.com/DIEALOG/DIEALOG-backend/config"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/internal/token"
	"github.com/gofiber/fiber/v3"
)

func JwtMiddleware(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{Message: "Authorization header is empty"})
	}
	t := strings.Split(authHeader, " ")
	if len(t) == 2 {
		authToken := t[1]
		authorized, err := token.IsAuthorized(authToken, config.E.AccessTokenSecret)
		if authorized {
			userID, err := token.ExtractIDFromToken(authToken, config.E.AccessTokenSecret)
			if err != nil {
				return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{Message: err.Error()})
			}
			c.Locals("id", userID)
			return c.Next()
		}
		return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{Message: err.Error()})
	}
	return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{Message: "Not authorized"})
}
