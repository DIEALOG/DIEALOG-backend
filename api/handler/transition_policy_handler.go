package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// FetchTransitionPolicy godoc
// @Summary      사후 전환 정책 조회
// @Description  현재 사용자의 사후 전환 정책을 조회합니다. 정책이 없으면 기본 정책을 생성합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} domain.UserTransitionPolicy "정책 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/policy [get]
func FetchTransitionPolicy(service domain.UserTransitionPolicyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		policy, err := service.Get(c.RequestCtx(), transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(policy)
	}
}

// UpdateTransitionPolicy godoc
// @Summary      사후 전환 정책 수정
// @Description  동의자 최소 수, 대기 기간, 생존 확인 주기를 수정합니다.
// @Tags         Transition
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.UpdateUserTransitionPolicyRequest true "수정할 정책"
// @Success      200 {object} domain.UserTransitionPolicy "정책 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 또는 정책 값"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/policy [patch]
func UpdateTransitionPolicy(service domain.UserTransitionPolicyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request domain.UpdateUserTransitionPolicyRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		policy, err := service.Update(c.RequestCtx(), &request, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(policy)
	}
}
