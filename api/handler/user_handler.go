package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// UpdateUser godoc
// @Summary      내 정보 수정
// @Description  현재 로그인된 사용자의 정보를 수정합니다.
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.User true "수정할 사용자 정보"
// @Success      200 {object} domain.User "사용자 정보 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /user/protected [put]
func UpdateUser(service domain.UserUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		var req domain.User
		err := c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		// 요청된 userID가 인증된 userID와 일치하는지 확인
		req.ID = *userID

		user, err := service.Update(ctx, &req)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(user)
	}
}

// DeleteUser godoc
// @Summary      내 계정 삭제
// @Description  현재 로그인된 사용자의 계정을 삭제합니다.
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} domain.SuccessResponse "계정 삭제 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /user/protected [delete]
func DeleteUser(service domain.UserUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		err := service.Delete(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{Message: "계정이 삭제되었습니다."})
	}
}