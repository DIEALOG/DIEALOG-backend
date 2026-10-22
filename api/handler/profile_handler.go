package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// FetchProfile godoc
// @Summary      내 프로필 조회
// @Description  현재 로그인된 사용자의 기본 프로필 및 등록된 이해관계자 정보를 조회합니다.
// @Tags         Profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} domain.Profile "프로필 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /profile/protected [get]
func FetchProfile(service domain.ProfileUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		profile, err := service.GetProfileByID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(profile)
	}
}

// UpdateTestament godoc
// @Summary      유언장 내용 수정
// @Description  사용자의 유언장(Testament) 내용을 등록 또는 수정합니다.
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.UpdateTestamentRequest true "수정할 유언장 내용"
// @Success      200 {object} domain.SuccessResponse "유언장 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /profile/protected/testament [put]
func UpdateTestament(service domain.ProfileUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		var req domain.UpdateTestamentRequest
		err := c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		err = service.UpdateTestament(ctx, userID, req.Testament)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Testament updated successfully",
		})
	}
}
