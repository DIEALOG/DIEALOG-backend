package handler

import (
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// GetNotifications
// @Summary 현재 사용자의 모든 알림 조회
// @Description 현재 로그인한 사용자의 모든 알림을 조회합니다.
// @Tags 알림
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.Notification "알림 목록"
// @Failure 401 {object} map[string]string "인증 실패"
// @Failure 500 {object} map[string]string "서버 내부 오류"
// @Security BearerAuth
// @Router /notification/protected [get]
func GetNotifications(service domain.NotificationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		notifications, err := service.ListByReceiverID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(200).JSON(notifications)
	}
}

// MarkNotificationAsRead
// @Summary 특정 알림 읽음 처리
// @Description 알림 ID를 이용하여 해당 알림을 읽음 상태로 변경합니다.
// @Tags 알림
// @Accept  json
// @Produce  json
// @Param id path int true "알림 ID"
// @Success 200 {object} map[string]string "성공 응답"
// @Failure 400 {object} map[string]string "잘못된 요청 (예: ID 변환 실패)"
// @Failure 401 {object} map[string]string "인증 실패"
// @Failure 500 {object} map[string]string "서버 내부 오류"
// @Security BearerAuth
// @Router /notification/protected/{id} [patch]
func MarkNotificationAsRead(service domain.NotificationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)
		notificationID, err := domain.StringToID(c.Params("id"))
		if err != nil {
			return RespondError(c, domain.NewBadRequestError(err))
		}

		err = service.MarkAsRead(ctx, &notificationID, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	}
}

// GetUnreadNotificationsCount
// @Summary 읽지 않은 알림 개수 조회
// @Description 현재 로그인한 사용자의 읽지 않은 알림 총 개수를 조회합니다.
// @Tags 알림
// @Accept  json
// @Produce  json
// @Success 200 {object} domain.NotificationCountResponse "읽지 않은 알림 개수"
// @Failure 401 {object} map[string]string "인증 실패"
// @Failure 500 {object} map[string]string "서버 내부 오류"
// @Security BearerAuth
// @Router /notification/protected/unread-count [get]
func GetUnreadNotificationsCount(service domain.NotificationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		count, err := service.CountUnreadByReceiverID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(200).JSON(&domain.NotificationCountResponse{Count: count})
	}
}

// DeleteNotification
// @Summary 특정 알림 삭제
// @Description 알림 ID를 이용하여 해당 알림을 삭제합니다.
// @Tags 알림
// @Accept  json
// @Produce  json
// @Param id path int true "알림 ID"
// @Success 200 {object} map[string]string "성공 응답"
// @Failure 400 {object} map[string]string "잘못된 요청 (예: ID 변환 실패)"
// @Failure 401 {object} map[string]string "인증 실패"
// @Failure 500 {object} map[string]string "서버 내부 오류"
// @Security BearerAuth
// @Router /notification/protected/{id} [delete]
func DeleteNotification(service domain.NotificationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)
		notificationID, err := domain.StringToID(c.Params("id"))
		if err != nil {
			return RespondError(c, domain.NewBadRequestError(err))
		}

		err = service.Delete(ctx, &notificationID, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	}
}
