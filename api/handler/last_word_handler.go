package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
	"github.com/janghanul090801/collections"
)

// CreateLastWord godoc
// @Summary      유언 메시지 생성
// @Description  새로운 유언 메시지를 생성하고 수신자 및 수신자 그룹을 지정합니다.
// @Tags         LastWord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.CreateLastWordRequest true "생성할 유언 메시지 정보"
// @Success      201 {object} domain.LastWord "유언 메시지 생성 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /last-word/protected [post]
func CreateLastWord(service domain.LastWordUseCase, stakeholderService domain.StakeholderUseCase, stakeholderGroupService domain.StakeholderGroupUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		var req domain.CreateLastWordRequest
		err := c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		userID := c.Locals("id").(*domain.ID)

		receivers, err := stakeholderService.ListByIDs(ctx, req.ReceiverIDs, userID)
		if err != nil {
			return RespondError(c, err)
		}

		receiverGroups, err := stakeholderGroupService.ListByIDs(ctx, req.ReceiverGroupIDs, userID)
		if err != nil {
			return RespondError(c, err)
		}

		lw, err := service.Create(ctx, &domain.LastWord{
			UserID:  *userID,
			Title:   req.Title,
			Message: req.Message,
			Receiver: collections.ToIterable(receivers).Map(func(r *domain.Stakeholder) domain.Stakeholder {
				return *r
			}),
			ReceiverGroups: collections.ToIterable(receiverGroups).Map(func(r *domain.StakeholderGroup) domain.StakeholderGroup {
				return *r
			}),
		})
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusCreated).JSON(lw)
	}
}

// FetchMyLastWords godoc
// @Summary      유언 메시지 목록 조회
// @Description  현재 사용자가 생성한 유언 메시지 목록을 조회합니다.
// @Tags         LastWord
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.LastWord "유언 메시지 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /last-word/protected [get]
func FetchMyLastWords(service domain.LastWordUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		userID := c.Locals("id").(*domain.ID)

		lastWords, err := service.ListByUserID(ctx, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(lastWords)
	}
}

// GetLastWord godoc
// @Summary      유언 메시지 상세 조회
// @Description  ID에 해당하는 유언 메시지의 상세 정보를 조회합니다.
// @Tags         LastWord
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "유언 메시지 ID (UUID)"
// @Success      200 {object} domain.LastWord "유언 메시지 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "유언 메시지를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /last-word/protected/{id} [get]
func GetLastWord(service domain.LastWordUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)
		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		lastWord, err := service.GetByID(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(lastWord)
	}
}

// UpdateLastWord godoc
// @Summary      유언 메시지 수정
// @Description  특정 유언 메시지의 제목, 내용, 수신자 및 수신자 그룹을 수정합니다.
// @Tags         LastWord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "유언 메시지 ID (UUID)"
// @Param        request body domain.UpdateLastWordRequest true "수정할 유언 메시지 정보"
// @Success      200 {object} domain.LastWord "유언 메시지 수정 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 파라미터 또는 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "유언 메시지를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /last-word/protected/{id} [put]
func UpdateLastWord(service domain.LastWordUseCase, stakeholderService domain.StakeholderUseCase, stakeholderGroupService domain.StakeholderGroupUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)
		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		var req domain.UpdateLastWordRequest
		err = c.Bind().Body(&req)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		receivers, err := stakeholderService.ListByIDs(ctx, req.ReceiverIDs, userID)
		if err != nil {
			return RespondError(c, err)
		}

		receiverGroups, err := stakeholderGroupService.ListByIDs(ctx, req.ReceiverGroupIDs, userID)
		if err != nil {
			return RespondError(c, err)
		}

		lw, err := service.Update(ctx, &id, &domain.LastWord{
			UserID:  *userID,
			Title:   req.Title,
			Message: req.Message,
			Receiver: collections.ToIterable(receivers).Map(func(r *domain.Stakeholder) domain.Stakeholder {
				return *r
			}),
			ReceiverGroups: collections.ToIterable(receiverGroups).Map(func(r *domain.StakeholderGroup) domain.StakeholderGroup {
				return *r
			}),
		}, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(lw)
	}
}

// DeleteLastWord godoc
// @Summary      유언 메시지 삭제
// @Description  특정 유언 메시지를 삭제합니다.
// @Tags         LastWord
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "유언 메시지 ID (UUID)"
// @Success      200 {object} domain.SuccessResponse "유언 메시지 삭제 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID 형식"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "유언 메시지를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /last-word/protected/{id} [delete]
func DeleteLastWord(service domain.LastWordUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()
		userID := c.Locals("id").(*domain.ID)
		idStr := c.Params("id")
		id, err := domain.StringToID(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		err = service.Delete(ctx, &id, userID)
		if err != nil {
			return RespondError(c, err)
		}

		return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
			Message: "LastWord deleted successfully",
		})
	}
}
