package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// FetchAliveChecks godoc
// @Summary      생존 확인 이력 조회
// @Description  현재 사용자의 생존 확인 일정과 응답 이력을 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.UserAliveCheck "생존 확인 이력 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/alive-checks [get]
func FetchAliveChecks(service domain.UserAliveCheckUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		items, err := service.ListByUserID(c.RequestCtx(), transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(items)
	}
}

// ScheduleAliveCheck godoc
// @Summary      다음 생존 확인 예약
// @Description  정책의 주기에 따라 다음 생존 확인 일정을 예약합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Success      201 {object} domain.UserAliveCheck "생존 확인 예약 성공"
// @Failure      400 {object} domain.ErrorResponse "이미 예약된 일정이 있거나 기한 초과 일정이 있음"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/alive-checks [post]
func ScheduleAliveCheck(service domain.UserAliveCheckUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		item, err := service.ScheduleNext(c.RequestCtx(), transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusCreated).JSON(item)
	}
}

// RespondAliveCheck godoc
// @Summary      생존 확인 응답
// @Description  생존 확인에 응답하고 다음 생존 확인 일정을 예약합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "생존 확인 ID (UUID)"
// @Success      200 {object} domain.UserAliveCheck "생존 확인 응답 성공"
// @Failure      400 {object} domain.ErrorResponse "이미 종료된 일정 또는 잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "생존 확인을 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/alive-checks/{id}/respond [post]
func RespondAliveCheck(service domain.UserAliveCheckUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid alive check id format")
		if err != nil {
			return RespondError(c, err)
		}

		item, err := service.Respond(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(item)
	}
}
