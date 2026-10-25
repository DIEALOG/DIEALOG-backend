package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// FetchTransitionConfirmers godoc
// @Summary      사후 전환 확인자 목록 조회
// @Description  현재 사용자가 초대한 사후 전환 확인자 목록을 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.UserTransitionConfirmer "확인자 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/confirmers [get]
func FetchTransitionConfirmers(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		items, err := service.ListByUserID(c.RequestCtx(), transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(items)
	}
}

// InviteTransitionConfirmer godoc
// @Summary      사후 전환 확인자 초대
// @Description  확인자를 초대하고 초대 토큰을 한 번만 반환합니다. 실제 이메일/SMS 발송은 아직 구현되지 않았습니다.
// @Tags         Transition
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.InviteTransitionConfirmerRequest true "초대할 확인자 정보"
// @Success      201 {object} domain.TransitionInvitationResponse "확인자 초대 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 또는 중복 확인자"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "이해관계자를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/confirmers [post]
func InviteTransitionConfirmer(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request domain.InviteTransitionConfirmerRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		result, err := service.Invite(c.RequestCtx(), &request, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusCreated).JSON(result)
	}
}

// ResendTransitionConfirmerInvitation godoc
// @Summary      사후 전환 확인자 초대 재발송
// @Description  기존 초대 토큰을 폐기하고 새 초대 토큰을 발급합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "확인자 ID (UUID)"
// @Success      200 {object} domain.TransitionInvitationResponse "초대 재발송 성공"
// @Failure      400 {object} domain.ErrorResponse "재발송할 수 없는 상태 또는 잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "확인자를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/confirmers/{id}/resend [post]
func ResendTransitionConfirmerInvitation(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid confirmer id format")
		if err != nil {
			return RespondError(c, err)
		}

		result, err := service.Resend(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// RevokeTransitionConfirmer godoc
// @Summary      사후 전환 확인자 철회
// @Description  확인자 지정을 철회하고 기존 초대/인증 토큰을 무효화합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "확인자 ID (UUID)"
// @Success      200 {object} domain.SuccessResponse "확인자 철회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "확인자를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/confirmers/{id} [delete]
func RevokeTransitionConfirmer(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid confirmer id format")
		if err != nil {
			return RespondError(c, err)
		}

		if err := service.Revoke(c.RequestCtx(), id, transitionUserID(c)); err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{Message: "confirmer revoked successfully"})
	}
}

// AcceptTransitionInvitation godoc
// @Summary      사후 전환 확인자 초대 수락
// @Description  요청 본문의 초대 토큰을 검증하고 확인자 인증 토큰을 발급합니다. 인증 토큰은 한 번만 반환됩니다.
// @Tags         Transition
// @Accept       json
// @Produce      json
// @Param        request body domain.TransitionInvitationTokenRequest true "초대 토큰"
// @Success      200 {object} domain.AcceptTransitionInvitationResponse "초대 수락 성공"
// @Failure      400 {object} domain.ErrorResponse "만료되었거나 유효하지 않은 초대"
// @Failure      404 {object} domain.ErrorResponse "초대를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/invitations/accept [post]
func AcceptTransitionInvitation(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request domain.TransitionInvitationTokenRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		result, err := service.Accept(c.RequestCtx(), request.Token)
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(result)
	}
}

// DeclineTransitionInvitation godoc
// @Summary      사후 전환 확인자 초대 거절
// @Description  요청 본문의 초대 토큰을 검증하고 초대를 거절합니다.
// @Tags         Transition
// @Accept       json
// @Produce      json
// @Param        request body domain.TransitionInvitationTokenRequest true "초대 토큰"
// @Success      200 {object} domain.SuccessResponse "초대 거절 성공"
// @Failure      400 {object} domain.ErrorResponse "만료되었거나 유효하지 않은 초대"
// @Failure      404 {object} domain.ErrorResponse "초대를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/invitations/decline [post]
func DeclineTransitionInvitation(service domain.UserTransitionConfirmerUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request domain.TransitionInvitationTokenRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		if err := service.Decline(c.RequestCtx(), request.Token); err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{Message: "invitation declined successfully"})
	}
}
