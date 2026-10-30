package handler

import (
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

// StartTransitionCase godoc
// @Summary      사후 전환 절차 생성
// @Description  수락된 확인자가 사후 전환 절차를 생성합니다. 사용자당 진행 중인 절차는 하나만 허용됩니다.
// @Tags         Transition
// @Produce      json
// @Security     ConfirmerAuth
// @Success      201 {object} domain.UserTransitionCase "전환 절차 생성 성공"
// @Failure      400 {object} domain.ErrorResponse "이미 진행 중인 절차가 있음"
// @Failure      401 {object} domain.ErrorResponse "확인자 인증 실패"
// @Failure      403 {object} domain.ErrorResponse "수락되지 않은 확인자"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/confirmer/cases [post]
func StartTransitionCase(service domain.UserTransitionCaseUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		item, err := service.Start(c.RequestCtx(), transitionConfirmer(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusCreated).JSON(item)
	}
}

// FetchTransitionCases godoc
// @Summary      사후 전환 절차 목록 조회
// @Description  현재 사용자의 사후 전환 절차 목록을 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.UserTransitionCase "전환 절차 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/cases [get]
func FetchTransitionCases(service domain.UserTransitionCaseUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		items, err := service.ListByUserID(c.RequestCtx(), transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(items)
	}
}

// FetchTransitionCase godoc
// @Summary      사후 전환 절차 상세 조회
// @Description  전환 절차 상태를 조회하고, 확정 조건이 충족되면 확정 처리합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "전환 절차 ID (UUID)"
// @Success      200 {object} domain.UserTransitionCase "전환 절차 상세 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "전환 절차를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/cases/{id} [get]
func FetchTransitionCase(service domain.UserTransitionCaseUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid transition case id format")
		if err != nil {
			return RespondError(c, err)
		}
		item, err := service.GetByID(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(item)
	}
}

// FetchConfirmerCases godoc
// @Summary      확인자용 전환 절차 목록 조회
// @Description  확인자 인증 토큰으로 자신이 확인하는 사용자의 진행 중인 전환 절차를 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     ConfirmerAuth
// @Success      200 {array} domain.UserTransitionCase "전환 절차 목록 조회 성공"
// @Failure      401 {object} domain.ErrorResponse "확인자 인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/confirmer/cases [get]
func FetchConfirmerCases(service domain.UserTransitionCaseUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		confirmer := transitionConfirmer(c)
		items, err := service.ListActiveByUserID(c.RequestCtx(), &confirmer.UserID)
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(items)
	}
}

// FetchConfirmerCase godoc
// @Summary      확인자용 전환 절차 상세 조회
// @Description  확인자 인증 토큰으로 전환 절차 상태를 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     ConfirmerAuth
// @Param        id path string true "전환 절차 ID (UUID)"
// @Success      200 {object} domain.UserTransitionCase "전환 절차 상세 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "확인자 인증 실패"
// @Failure      404 {object} domain.ErrorResponse "전환 절차를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/confirmer/cases/{id} [get]
func FetchConfirmerCase(service domain.UserTransitionCaseUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid transition case id format")
		if err != nil {
			return RespondError(c, err)
		}
		confirmer := transitionConfirmer(c)
		item, err := service.GetByID(c.RequestCtx(), id, &confirmer.UserID)
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(item)
	}
}

// CancelTransitionCase godoc
// @Summary      사후 전환 절차 취소
// @Description  확정되지 않은 전환 절차를 취소합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "전환 절차 ID (UUID)"
// @Success      200 {object} domain.UserTransitionCase "전환 절차 취소 성공"
// @Failure      400 {object} domain.ErrorResponse "이미 확정된 절차 또는 잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "전환 절차를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/cases/{id}/cancel [post]
func CancelTransitionCase(service domain.UserTransitionCaseUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid transition case id format")
		if err != nil {
			return RespondError(c, err)
		}
		item, err := service.Cancel(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(item)
	}
}

// SubmitTransitionConfirmation godoc
// @Summary      사후 전환 동의/반대 제출
// @Description  확인자가 전환 절차에 동의, 반대, 보류 의사를 제출합니다. 동의 기준 충족 전에는 변경할 수 있습니다.
// @Tags         Transition
// @Accept       json
// @Produce      json
// @Security     ConfirmerAuth
// @Param        id path string true "전환 절차 ID (UUID)"
// @Param        request body domain.SubmitTransitionConfirmationRequest true "확인 의사"
// @Success      200 {object} domain.UserTransitionConfirmation "확인 의사 제출 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청 또는 변경할 수 없는 상태"
// @Failure      401 {object} domain.ErrorResponse "확인자 인증 실패"
// @Failure      403 {object} domain.ErrorResponse "수락되지 않은 확인자"
// @Failure      404 {object} domain.ErrorResponse "전환 절차를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/confirmer/cases/{id}/confirmations [post]
func SubmitTransitionConfirmation(service domain.UserTransitionConfirmationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid transition case id format")
		if err != nil {
			return RespondError(c, err)
		}

		var request domain.SubmitTransitionConfirmationRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}

		item, err := service.Submit(c.RequestCtx(), id, transitionConfirmer(c), &request)
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(item)
	}
}

// FetchTransitionConfirmations godoc
// @Summary      사후 전환 확인 의사 목록 조회
// @Description  전환 절차에 제출된 확인 의사 목록을 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "전환 절차 ID (UUID)"
// @Success      200 {array} domain.UserTransitionConfirmation "확인 의사 목록 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/cases/{id}/confirmations [get]
func FetchTransitionConfirmations(service domain.UserTransitionConfirmationUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid transition case id format")
		if err != nil {
			return RespondError(c, err)
		}
		items, err := service.ListByCaseID(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(items)
	}
}

// FetchAccountHandlingTasks godoc
// @Summary      계정 처리 작업 목록 조회
// @Description  확정된 전환 절차의 계정별 처리 작업과 이벤트 이력을 조회합니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "전환 절차 ID (UUID)"
// @Success      200 {array} domain.AccountHandlingTask "계정 처리 작업 목록 조회 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "전환 절차를 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/cases/{id}/tasks [get]
func FetchAccountHandlingTasks(service domain.AccountHandlingUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid transition case id format")
		if err != nil {
			return RespondError(c, err)
		}
		items, err := service.ListByCaseID(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(items)
	}
}

// RequestAccountHandlingTask godoc
// @Summary      계정 처리 요청 토큰 발급
// @Description  계정에 지정된 관계자에게 전달할 사후 처리 작업 요청 토큰을 발급합니다. 요청할 때마다 계정의 현재 관계자를 추가로 반영하고 기존 관계자는 유지합니다. 요청 토큰은 이 응답에서만 반환됩니다.
// @Tags         Transition
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "계정 처리 작업 ID (UUID)"
// @Success      200 {object} domain.AccountHandlingRequestResponse "계정 처리 요청 토큰 발급 결과"
// @Failure      400 {object} domain.ErrorResponse "요청할 수 없는 상태, 관계자 없음 또는 잘못된 ID"
// @Failure      401 {object} domain.ErrorResponse "인증 실패"
// @Failure      404 {object} domain.ErrorResponse "작업을 찾을 수 없음"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/protected/tasks/{id}/request [post]
func RequestAccountHandlingTask(service domain.AccountHandlingUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := transitionParamID(c, "id", "invalid task id format")
		if err != nil {
			return RespondError(c, err)
		}
		item, err := service.Request(c.RequestCtx(), id, transitionUserID(c))
		if err != nil {
			return RespondError(c, err)
		}
		return c.Status(http.StatusOK).JSON(item)
	}
}

// SubmitAccountHandlingResult godoc
// @Summary      계정 처리 성공 여부 보고
// @Description  관계자가 계정 처리 요청 토큰으로 실제 처리 성공 또는 실패 여부를 보고합니다.
// @Tags         Transition
// @Accept       json
// @Produce      json
// @Param        request body domain.SubmitAccountHandlingResultRequest true "처리 결과와 요청 토큰"
// @Success      200 {object} domain.AccountHandlingTask "계정 처리 결과 보고 성공"
// @Failure      400 {object} domain.ErrorResponse "잘못된 요청, 만료된 토큰 또는 처리할 수 없는 상태"
// @Failure      500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router       /transition/stakeholder-tasks/result [post]
func SubmitAccountHandlingResult(service domain.AccountHandlingUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request domain.SubmitAccountHandlingResultRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{Message: err.Error()})
		}
		item, err := service.SubmitResult(c.RequestCtx(), &request)
		if err != nil {
			return RespondError(c, err)
		}
		item.Assignees = nil
		item.Events = nil
		return c.Status(http.StatusOK).JSON(item)
	}
}
