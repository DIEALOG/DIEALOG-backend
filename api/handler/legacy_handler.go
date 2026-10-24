package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

type UpdateLegacyRequest struct {
	Name string `form:"name" binding:"required" json:"name"`
}

// CreateLegacy godoc
// @Summary      레거시 생성
// @Description  새로운 레거시(디지털 유산 그룹) 항목을 생성합니다.
// @Tags         Legacy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.Legacy true "생성할 레거시 정보"
// @Success      201 {object} domain.LegacyResponse "레거시 생성 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected [post]
func CreateLegacy(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var legacy domain.Legacy
		err := c.Bind().Body(&legacy)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		created, err := service.Create(ctx, &legacy, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(
			domain.ToLegacyResponse(*created),
		)
	}
}

// FetchLegacies godoc
// @Summary      레거시 목록 조회
// @Description  현재 사용자의 레거시 목록을 조회합니다.
// @Tags         Legacy
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.LegacyResponse "레거시 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected [get]
func FetchLegacies(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		legacies, err := service.ListByUserID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		response := make([]domain.LegacyResponse, 0, len(legacies))
		for _, legacy := range legacies {
			response = append(response, domain.ToLegacyResponse(*legacy))
		}

		return c.Status(http.StatusOK).JSON(response)
	}
}

// GetLegacy godoc
// @Summary      레거시 상세 조회
// @Description  ID에 해당하는 레거시의 상세 정보 및 하위 계정 목록을 조회합니다.
// @Tags         Legacy
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "레거시 ID (UUID)"
// @Success      200 {object} domain.LegacyResponse "레거시 상세 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "레거시를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected/{id} [get]
func GetLegacy(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		id, err := domain.StringToID(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.ErrorResponse{Message: "invalid id format"},
			)
		}

		legacy, err := service.GetByID(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(
			domain.ToLegacyResponse(*legacy),
		)
	}
}

// UpdateLegacy godoc
// @Summary      레거시 정보 수정
// @Description  특정 레거시의 이름을 수정합니다.
// @Tags         Legacy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "레거시 ID (UUID)"
// @Param        request body UpdateLegacyRequest true "수정할 레거시 정보"
// @Success      200 {object} domain.LegacyResponse "레거시 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터 또는 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "레거시를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected/{id} [put]
func UpdateLegacy(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		var req UpdateLegacyRequest
		err = c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		updated, err := service.Update(ctx, &id, req.Name, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(
			domain.ToLegacyResponse(*updated),
		)
	}
}

// DeleteLegacy godoc
// @Summary      레거시 삭제
// @Description  특정 레거시 및 연관된 데이터를 삭제합니다.
// @Tags         Legacy
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "레거시 ID (UUID)"
// @Success      200 {object} domain.SuccessResponse "레거시 삭제 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected/{id} [delete]
func DeleteLegacy(service domain.LegacyUseCase) fiber.Handler {
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
			Message: "Legacy deleted successfully",
		})
	}
}

// CreateAccount godoc
// @Summary      레거시 하위 계정 등록
// @Description  특정 레거시에 디지털 서비스 계정 정보 및 사후 처리 방침을 등록합니다.
// @Tags         Legacy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "레거시 ID (UUID)"
// @Param        request body domain.Account true "등록할 계정 정보"
// @Success      201 {object} domain.AccountDetailResponse "계정 등록 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터 또는 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected/{id}/account [post]
func CreateAccount(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		legacyIDStr := c.Params("id")
		legacyID, err := domain.StringToID(legacyIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid legacy id format"})
		}

		var req domain.CreateAccountRequest
		err = c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		created, err := service.CreateAccount(ctx, &domain.Account{
			Type:           req.Type,
			Name:           req.Name,
			AccountID:      req.AccountID,
			DataLocation:   req.DataLocation,
			StakeholderIDs: req.StakeholderIDs,
			Action:         req.Action,
		}, req.AccountPassword, &legacyID, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(domain.ToAccountDetailResponse(*created, req.AccountPassword))
	}
}

// GetAccount godoc
// @Summary      특정 계정 detail 정보
// @Description  특정 계정의 상세 정보(비밀번호 포함) 을 조회합니다
// @Tags         Legacy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "레거시 ID (UUID)"
// @Param        accountId path string true "계정 ID (UUID)"
// @Success      200 {object} domain.AccountDetailResponse "계정 정보"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터 또는 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected/{id}/account/{accountId} [get]
func GetAccount(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)
		idStr := c.Params("accountId")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid id format"})
		}

		acc, pwd, err := service.GetAccountByID(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.ToAccountDetailResponse(*acc, pwd))
	}
}

// DeleteAccount godoc
// @Summary      레거시 하위 계정 삭제
// @Description  특정 레거시에 등록된 디지털 서비스 계정을 삭제합니다.
// @Tags         Legacy
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "레거시 ID (UUID)"
// @Param        accountId path string true "계정 ID (UUID)"
// @Success      200 {object} domain.SuccessResponse "계정 삭제 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /legacy/protected/{id}/account/{accountId} [delete]
func DeleteAccount(service domain.LegacyUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)

		legacyIDStr := c.Params("id")
		legacyID, err := domain.StringToID(legacyIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid legacy id format"})
		}

		accountIDStr := c.Params("accountId")
		accountID, err := domain.StringToID(accountIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: "invalid account id format"})
		}

		err = service.DeleteAccount(ctx, &accountID, &legacyID, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "Account deleted successfully",
		})
	}
}
