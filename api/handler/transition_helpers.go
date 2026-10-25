package handler

import (
	"errors"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func transitionUserID(c fiber.Ctx) *domain.ID {
	return c.Locals("id").(*domain.ID)
}

func transitionConfirmer(c fiber.Ctx) *domain.UserTransitionConfirmer {
	return c.Locals("confirmer").(*domain.UserTransitionConfirmer)
}

func transitionParamID(c fiber.Ctx, name, message string) (*domain.ID, error) {
	id, err := domain.StringToID(c.Params(name))
	if err != nil {
		return nil, domain.NewBadRequestError(errors.New(message))
	}
	return &id, nil
}
