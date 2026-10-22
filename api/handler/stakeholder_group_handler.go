package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// CreateStakeholderGroup godoc
// @Summary      관계자 그룹 등록
// @Description  이해관계자들을 묶어 관리할 수 있는 새로운 그룹을 등록합니다.
// @Tags         Stakeholder Group
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.StakeholderGroup true "등록할 관계자 그룹 정보"
// @Success      201 {object} domain.StakeholderGroup "관계자 그룹 등록 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/group [post]
func CreateStakeholderGroup(service domain.StakeholderGroupUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var group domain.StakeholderGroup
		err := c.Bind().Body(&group)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		created, err := service.Create(ctx, &group, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(created)
	}
}

// FetchStakeholderGroups godoc
// @Summary      관계자 그룹 목록 조회
// @Description  현재 사용자가 등록한 모든 관계자 그룹 목록을 조회합니다.
// @Tags         Stakeholder Group
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.StakeholderGroup "관계자 그룹 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/group [get]
func FetchStakeholderGroups(service domain.StakeholderGroupUseCase) fiber.Handler {
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

// GetStakeholderGroup godoc
// @Summary      관계자 그룹 상세 조회
// @Description  ID에 해당하는 특정 관계자 그룹의 상세 정보를 조회합니다.
// @Tags         Stakeholder Group
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "관계자 그룹 ID (UUID)"
// @Success      200 {object} domain.StakeholderGroup "그룹 상세 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "그룹을 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/group/{id} [get]
func GetStakeholderGroup(service domain.StakeholderGroupUseCase) fiber.Handler {
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

// UpdateStakeholderGroup godoc
// @Summary      관계자 그룹 정보 수정
// @Description  특정 관계자 그룹의 이름 및 속한 관계자 목록을 수정합니다.
// @Tags         Stakeholder Group
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "관계자 그룹 ID (UUID)"
// @Param        request body domain.StakeholderGroup true "수정할 관계자 그룹 정보"
// @Success      200 {object} domain.StakeholderGroup "그룹 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터 또는 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "그룹을 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/group/{id} [put]
func UpdateStakeholderGroup(service domain.StakeholderGroupUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		var group domain.StakeholderGroup
		err = c.Bind().Body(&group)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		updated, err := service.Update(ctx, &id, &group, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(updated)
	}
}

// DeleteStakeholderGroup godoc
// @Summary      관계자 그룹 삭제
// @Description  특정 관계자 그룹을 삭제합니다.
// @Tags         Stakeholder Group
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "관계자 그룹 ID (UUID)"
// @Success      200 {object} domain.SuccessResponse "그룹 삭제 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /stakeholder/protected/group/{id} [delete]
func DeleteStakeholderGroup(service domain.StakeholderGroupUseCase) fiber.Handler {
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
			Message: "Stakeholder group deleted successfully",
		})
	}
}
