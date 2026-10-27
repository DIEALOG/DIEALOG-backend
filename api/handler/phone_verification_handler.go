package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// SendPhoneVerification godoc
// @Summary      전화번호 인증번호 발송
// @Description  지정한 전화번호로 SMS 인증번호 발송을 요청합니다.
// @Tags         PhoneVerification
// @Accept       json
// @Produce      json
// @Param        request body domain.SendPhoneVerificationRequest true "전화번호 인증 요청"
// @Success      200 {object} domain.SuccessResponse "인증번호 발송 요청 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      500 {object} domain.ErrorResponse "인증번호 발송 실패"
// @Router       /phone-verification/send [post]
func SendPhoneVerification(service domain.PhoneVerificationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var request domain.SendPhoneVerificationRequest

		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.ErrorResponse{Message: err.Error()},
			)
		}

		if err := service.SendCode(ctx, request.PhoneNumber); err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Verification code sent",
		})
	}
}

// VerifyPhoneVerification godoc
// @Summary      전화번호 인증번호 검증
// @Description  입력한 인증번호를 검증합니다.
// @Tags         PhoneVerification
// @Accept       json
// @Produce      json
// @Param        request body domain.VerifyPhoneVerificationRequest true "전화번호 및 인증번호"
// @Success      200 {object} domain.SuccessResponse "전화번호 인증 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "인증번호 불일치"
// @Failure      500 {object} domain.ErrorResponse "인증 처리 실패"
// @Router       /phone-verification/verify [post]
func VerifyPhoneVerification(service domain.PhoneVerificationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var request domain.VerifyPhoneVerificationRequest

		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.ErrorResponse{Message: err.Error()},
			)
		}

		if err := service.VerifyCode(
			ctx,
			request.PhoneNumber,
			request.Code,
		); err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Phone number verified successfully",
		})
	}
}
