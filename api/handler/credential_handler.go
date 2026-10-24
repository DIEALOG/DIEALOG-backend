package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func StorePassword(service domain.CredentialUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req domain.StoreCredentialRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "잘못된 요청 형식입니다."})
		}

		userID, ok := c.Locals("userID").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{Message: "인증 권한이 없습니다."})
		}

		if err := service.StorePassword(c.Context(), userID, &req); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{Message: "비밀번호가 성공적으로 저장되었습니다."})
	}
}

func GetPassword(service domain.CredentialUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req domain.GetPasswordRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "잘못된 요청 형식입니다."})
		}

		userID, ok := c.Locals("userID").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{Message: "인증 권한이 없습니다."})
		}

		password, err := service.GetPassword(c.Context(), userID, &req)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"account_name": req.AccountName,
			"password":     password,
		})
	}
}
