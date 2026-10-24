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

// UpdateProfileImage godoc
// @Summary      내 프로필 이미지 수정
// @Description  현재 로그인된 사용자의 프로필 이미지를 업로드하고 수정합니다.
// @Tags         Profile
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        profile_image formData file true "프로필 이미지 파일"
// @Success      200 {object} domain.Profile "프로필 이미지 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 또는 이미지 누락"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /profile/protected/image [put]
func UpdateProfileImage(service domain.ProfileUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		userID := c.Locals("id").(*domain.ID)

		file, err := c.FormFile("profile_image")
		if err != nil {
			return fiber.NewError(
				fiber.StatusBadRequest,
				"프로필 이미지를 첨부해주세요.",
			)
		}

		profile, err := service.UpdateProfileImage(ctx, userID, file)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(profile)
	}
}
