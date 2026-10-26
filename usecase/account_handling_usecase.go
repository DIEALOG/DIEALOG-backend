package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/internal/securetoken"
)

const accountHandlingRequestTTL = 7 * 24 * time.Hour

type accountHandlingUseCase struct {
	taskRepository   domain.AccountHandlingTaskRepository
	eventRepository  domain.AccountHandlingEventRepository
	legacyRepository domain.LegacyRepository
	caseUseCase      domain.UserTransitionCaseUseCase
	contextTimeout   time.Duration
}

func NewAccountHandlingUseCase(
	taskRepository domain.AccountHandlingTaskRepository,
	eventRepository domain.AccountHandlingEventRepository,
	legacyRepository domain.LegacyRepository,
	caseUseCase domain.UserTransitionCaseUseCase,
	timeout time.Duration,
) domain.AccountHandlingUseCase {
	return &accountHandlingUseCase{
		taskRepository:   taskRepository,
		eventRepository:  eventRepository,
		legacyRepository: legacyRepository,
		caseUseCase:      caseUseCase,
		contextTimeout:   timeout,
	}
}

func (u *accountHandlingUseCase) ListByCaseID(c context.Context, caseID *domain.ID, userID *domain.ID) ([]*domain.AccountHandlingTask, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	if _, err := u.caseUseCase.Evaluate(ctx, caseID, userID); err != nil {
		return nil, err
	}

	items, err := u.taskRepository.FindAllByCaseID(ctx, caseID, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	return items, nil
}

func (u *accountHandlingUseCase) Request(c context.Context, taskID *domain.ID, userID *domain.ID) (*domain.AccountHandlingRequestResponse, error) {
	if taskID == nil || userID == nil {
		return nil, domain.NewBadRequestError(errors.New("task id and user id are required"))
	}

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	task, err := u.taskRepository.FindByID(ctx, taskID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	transitionCase, err := u.caseUseCase.GetByID(ctx, &task.TransitionCaseID, userID)
	if err != nil {
		return nil, err
	}
	if transitionCase.Status != domain.UserTransitionCaseStatusConfirmed {
		return nil, domain.NewBadRequestError(errors.New("transition case is not confirmed"))
	}

	if task.Status == domain.AccountHandlingTaskStatusCompleted {
		return &domain.AccountHandlingRequestResponse{Task: task}, nil
	}
	if task.Status == domain.AccountHandlingTaskStatusSkipped {
		return nil, domain.NewBadRequestError(errors.New("skipped account handling task cannot be requested"))
	}

	// Refresh the assignee snapshot additively from the account's current
	// stakeholders. Existing assignees are retained even if the account's
	// stakeholder list changes later.
	account, err := u.legacyRepository.FindAccountById(ctx, &task.AccountID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	currentAssignees := stakeholderAssignees(account)
	if len(currentAssignees) == 0 && len(task.Assignees) == 0 {
		return nil, domain.NewBadRequestError(errors.New("account has no stakeholders assigned"))
	}
	if len(currentAssignees) > 0 {
		task, err = u.taskRepository.EnsureAssignees(ctx, &task.ID, userID, currentAssignees)
		if err != nil {
			return nil, domain.NewInternalServerError(err)
		}
	}

	now := time.Now()
	expiresAt := now.Add(accountHandlingRequestTTL)
	requested := make([]domain.AccountHandlingTaskAssignee, 0, len(task.Assignees))
	requests := make([]domain.AccountHandlingRequest, 0, len(task.Assignees))
	for i := range task.Assignees {
		assignee := task.Assignees[i]
		if assignee.Status == domain.AccountHandlingTaskAssigneeStatusConfirmed {
			continue
		}

		token, err := securetoken.Generate()
		if err != nil {
			return nil, domain.NewInternalServerError(err)
		}
		assignee.Status = domain.AccountHandlingTaskAssigneeStatusPending
		assignee.RequestedAt = &now
		assignee.RespondedAt = nil
		assignee.ResponseMessage = nil
		assignee.RequestTokenHash = securetoken.Hash(token)
		assignee.RequestTokenExpiresAt = &expiresAt
		requested = append(requested, assignee)
		requests = append(requests, domain.AccountHandlingRequest{
			AssigneeID:    assignee.ID,
			StakeholderID: assignee.StakeholderID,
			Name:          assignee.Name,
			Email:         assignee.Email,
			Phone:         assignee.Phone,
			Token:         token,
			ExpiresAt:     expiresAt,
		})
	}

	if len(requests) == 0 {
		return &domain.AccountHandlingRequestResponse{Task: task}, nil
	}

	if task.StartedAt == nil {
		task.StartedAt = &now
	}
	task.Status = domain.AccountHandlingTaskStatusInProgress
	task.CompletedAt = nil
	task.FailureReason = nil

	updated, err := u.taskRepository.SaveRequest(ctx, task, requested, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return nil, domain.NewBadRequestError(errors.New("account handling task cannot be requested in its current state"))
		}
		return nil, domain.NewInternalServerError(err)
	}
	if updated.Status == domain.AccountHandlingTaskStatusCompleted {
		// A concurrent result settled the task before the request was stored.
		// Do not report that tokens were issued.
		return &domain.AccountHandlingRequestResponse{Task: updated}, nil
	}

	message := fmt.Sprintf("account handling request token issued for %d stakeholder(s)", len(requests))
	if updated.RequestCount > 1 {
		message = fmt.Sprintf("account handling request token re-issued for %d stakeholder(s)", len(requests))
	}
	if _, err := u.eventRepository.Create(ctx, &domain.AccountHandlingEvent{
		TaskID:    updated.ID,
		EventType: domain.AccountHandlingEventTypeRequestIssued,
		Message:   &message,
	}, userID); err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	final, err := u.taskRepository.FindByID(ctx, &updated.ID, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.AccountHandlingRequestResponse{
		Task:     final,
		Requests: requests,
	}, nil
}

func (u *accountHandlingUseCase) SubmitResult(c context.Context, request *domain.SubmitAccountHandlingResultRequest) (*domain.AccountHandlingTask, error) {
	if request == nil {
		return nil, domain.NewBadRequestError(errors.New("request is required"))
	}
	token := strings.TrimSpace(request.Token)
	if token == "" {
		return nil, domain.NewBadRequestError(errors.New("account handling request token is required"))
	}
	if request.Success == nil {
		return nil, domain.NewBadRequestError(errors.New("success is required"))
	}

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	task, assignee, err := u.taskRepository.FindByAssigneeTokenHash(ctx, securetoken.Hash(token))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewBadRequestError(errors.New("invalid account handling request token"))
		}
		return nil, domain.NewInternalServerError(err)
	}

	// A stakeholder may retry after a network failure. Once this token has
	// already been used, replaying the same result returns the settled task
	// instead of creating a second event or changing the outcome.
	if assignee.Status == domain.AccountHandlingTaskAssigneeStatusConfirmed && task.Status == domain.AccountHandlingTaskStatusCompleted {
		if *request.Success {
			return task, nil
		}
		return nil, domain.NewBadRequestError(errors.New("account handling result has already been reported"))
	}
	if assignee.Status == domain.AccountHandlingTaskAssigneeStatusFailed && task.Status == domain.AccountHandlingTaskStatusFailed {
		if !*request.Success {
			return task, nil
		}
		return nil, domain.NewBadRequestError(errors.New("account handling result has already been reported"))
	}

	if task.Status != domain.AccountHandlingTaskStatusInProgress {
		return nil, domain.NewBadRequestError(errors.New("account handling task is not awaiting a stakeholder result"))
	}
	if assignee.Status != domain.AccountHandlingTaskAssigneeStatusPending {
		return nil, domain.NewBadRequestError(errors.New("account handling request is no longer active"))
	}
	now := time.Now()
	if assignee.RequestTokenExpiresAt == nil || !now.Before(*assignee.RequestTokenExpiresAt) {
		return nil, domain.NewBadRequestError(errors.New("account handling request token expired"))
	}

	message := normalizeOptionalString(request.Message)
	assignee.RespondedAt = &now
	assignee.ResponseMessage = message

	if *request.Success {
		assignee.Status = domain.AccountHandlingTaskAssigneeStatusConfirmed
		task.Status = domain.AccountHandlingTaskStatusCompleted
		task.CompletedAt = &now
		task.FailureReason = nil
	} else {
		assignee.Status = domain.AccountHandlingTaskAssigneeStatusFailed
		task.Status = domain.AccountHandlingTaskStatusFailed
		task.CompletedAt = nil
		reason := "stakeholder reported account handling failure"
		if message != nil && *message != "" {
			reason = *message
		}
		task.FailureReason = &reason
	}

	updated, err := u.taskRepository.SaveResult(ctx, task, assignee, &task.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return nil, domain.NewBadRequestError(errors.New("account handling task state changed; refresh and retry"))
		}
		return nil, domain.NewInternalServerError(err)
	}

	persistCtx, persistCancel := context.WithTimeout(context.WithoutCancel(c), u.contextTimeout)
	defer persistCancel()

	resultType := domain.AccountHandlingEventTypeStakeholderConfirmed
	resultMessage := fmt.Sprintf("stakeholder %s confirmed account handling success", assignee.Name)
	if !*request.Success {
		resultType = domain.AccountHandlingEventTypeStakeholderReportedFailed
		resultMessage = fmt.Sprintf("stakeholder %s reported account handling failure", assignee.Name)
		if message != nil && *message != "" {
			resultMessage = fmt.Sprintf("stakeholder %s reported account handling failure: %s", assignee.Name, *message)
		}
	}
	if _, err := u.eventRepository.Create(persistCtx, &domain.AccountHandlingEvent{
		TaskID:    updated.ID,
		EventType: resultType,
		Message:   &resultMessage,
	}, &task.UserID); err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	taskEventType := domain.AccountHandlingEventTypeTaskCompleted
	taskEventMessage := "account handling task completed by stakeholder confirmation"
	if !*request.Success {
		taskEventType = domain.AccountHandlingEventTypeTaskFailed
		taskEventMessage = "account handling task failed by stakeholder report"
	}
	if _, err := u.eventRepository.Create(persistCtx, &domain.AccountHandlingEvent{
		TaskID:    updated.ID,
		EventType: taskEventType,
		Message:   &taskEventMessage,
	}, &task.UserID); err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	final, err := u.taskRepository.FindByID(persistCtx, &updated.ID, &task.UserID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	return final, nil
}

func stakeholderAssignees(account *domain.Account) []domain.AccountHandlingTaskAssignee {
	if account == nil {
		return nil
	}

	seen := make(map[domain.ID]struct{})
	assignees := make([]domain.AccountHandlingTaskAssignee, 0)
	add := func(stakeholder domain.Stakeholder) {
		if _, ok := seen[stakeholder.ID]; ok {
			return
		}
		seen[stakeholder.ID] = struct{}{}
		assignees = append(assignees, domain.AccountHandlingTaskAssignee{
			StakeholderID: stakeholder.ID,
			Name:          stakeholder.Name,
			Email:         stakeholder.Email,
			Phone:         stakeholder.Phone,
		})
	}

	for _, stakeholder := range account.Stakeholders {
		add(stakeholder)
	}
	for _, group := range account.StakeholderGroups {
		for _, stakeholder := range group.Stakeholders {
			add(stakeholder)
		}
	}

	sort.Slice(assignees, func(i, j int) bool {
		if assignees[i].Name == assignees[j].Name {
			return assignees[i].StakeholderID.String() < assignees[j].StakeholderID.String()
		}
		return assignees[i].Name < assignees[j].Name
	})
	return assignees
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}
