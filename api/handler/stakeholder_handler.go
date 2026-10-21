package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
)

// CreateStakeholder godoc
// @Summary      관계자 등록
// @Description  사후 디지털 유산 전달 또는 연락 대상이 될 관계자(이해관계자) 정보를 등록합니다.
// @Tags         Stakeholder
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.Stakeholder true "등록할 관계자 정보"
// @Success      201 {object} domain.Stakeholder "관계자 등록 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected [post]
func CreateStakeholder(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var stakeholder domain.Stakeholder
		err := c.Bind().Body(&stakeholder)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		created, err := service.Create(ctx, &stakeholder, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(created)
	}
}

// FetchStakeholders godoc
// @Summary      관계자 목록 조회
// @Description  현재 사용자가 등록한 모든 관계자(이해관계자) 목록을 조회합니다.
// @Tags         Stakeholder
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.Stakeholder "관계자 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected [get]
func FetchStakeholders(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		list, err := service.ListByUserID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(list)
	}
}

// GetStakeholder godoc
// @Summary      관계자 상세 조회
// @Description  ID에 해당하는 특정 관계자의 상세 정보를 조회합니다.
// @Tags         Stakeholder
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "관계자 ID (UUID)"
// @Success      200 {object} domain.Stakeholder "관계자 상세 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "관계자를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/{id} [get]
func GetStakeholder(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		item, err := service.GetByID(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(item)
	}
}

// UpdateStakeholder godoc
// @Summary      관계자 정보 수정
// @Description  특정 관계자의 연락처, 이름, 관계 등의 정보를 수정합니다.
// @Tags         Stakeholder
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "관계자 ID (UUID)"
// @Param        request body domain.Stakeholder true "수정할 관계자 정보"
// @Success      200 {object} domain.Stakeholder "관계자 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터 또는 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "관계자를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/{id} [put]
func UpdateStakeholder(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		var stakeholder domain.Stakeholder
		err = c.Bind().Body(&stakeholder)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		updated, err := service.Update(ctx, &id, &stakeholder, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(updated)
	}
}

// DeleteStakeholder godoc
// @Summary      관계자 삭제
// @Description  특정 관계자 정보를 삭제합니다.
// @Tags         Stakeholder
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "관계자 ID (UUID)"
// @Success      200 {object} domain.SuccessResponse "관계자 삭제 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/{id} [delete]
func DeleteStakeholder(service domain.StakeholderUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		err = service.Delete(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Stakeholder deleted successfully",
		})
	}
}
