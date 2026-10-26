package repository

import (
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

func toDomainUserTransitionPolicy(entity *ent.UserTransitionPolicy) *domain.UserTransitionPolicy {
	item := &domain.UserTransitionPolicy{
		ID:                     entity.ID,
		MinimumConfirmers:      entity.MinimumConfirmers,
		WaitingPeriodDays:      entity.WaitingPeriodDays,
		AliveCheckIntervalDays: entity.AliveCheckIntervalDays,
		CreatedAt:              entity.CreatedAt,
		UpdatedAt:              entity.UpdatedAt,
	}
	if entity.Edges.Owner != nil {
		item.UserID = entity.Edges.Owner.ID
	}
	return item
}

func toDomainUserTransitionConfirmer(entity *ent.UserTransitionConfirmer) *domain.UserTransitionConfirmer {
	item := &domain.UserTransitionConfirmer{
		ID:                 entity.ID,
		Name:               entity.Name,
		Email:              entity.Email,
		Phone:              entity.Phone,
		Status:             domain.UserTransitionConfirmerStatus(entity.Status),
		InviteExpiresAt:    entity.InviteExpiresAt,
		AuthTokenExpiresAt: entity.AuthTokenExpiresAt,
		InvitedAt:          entity.InvitedAt,
		AcceptedAt:         entity.AcceptedAt,
		RevokedAt:          entity.RevokedAt,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
	if entity.Edges.Owner != nil {
		item.UserID = entity.Edges.Owner.ID
	}
	if entity.Edges.Stakeholder != nil {
		id := entity.Edges.Stakeholder.ID
		item.StakeholderID = &id
	}
	if entity.InviteTokenHash != nil {
		item.InviteTokenHash = *entity.InviteTokenHash
	}
	if entity.AuthTokenHash != nil {
		item.AuthTokenHash = *entity.AuthTokenHash
	}
	return item
}

func toDomainUserAliveCheck(entity *ent.UserAliveCheck) *domain.UserAliveCheck {
	item := &domain.UserAliveCheck{
		ID:          entity.ID,
		Sequence:    entity.Sequence,
		ScheduledAt: entity.ScheduledAt,
		DueAt:       entity.DueAt,
		RespondedAt: entity.RespondedAt,
		Status:      domain.UserAliveCheckStatus(entity.Status),
		CreatedAt:   entity.CreatedAt,
	}
	if entity.Edges.Owner != nil {
		item.UserID = entity.Edges.Owner.ID
	}
	return item
}

func toDomainUserTransitionCase(entity *ent.UserTransitionCase) *domain.UserTransitionCase {
	item := &domain.UserTransitionCase{
		ID:                        entity.ID,
		Status:                    domain.UserTransitionCaseStatus(entity.Status),
		MinimumConfirmersSnapshot: entity.MinimumConfirmersSnapshot,
		WaitingPeriodDaysSnapshot: entity.WaitingPeriodDaysSnapshot,
		RequestedAt:               entity.RequestedAt,
		WaitingStartedAt:          entity.WaitingStartedAt,
		EligibleAt:                entity.EligibleAt,
		ConfirmedAt:               entity.ConfirmedAt,
		CancelledAt:               entity.CancelledAt,
		CreatedAt:                 entity.CreatedAt,
		UpdatedAt:                 entity.UpdatedAt,
	}
	if entity.Edges.Owner != nil {
		item.UserID = entity.Edges.Owner.ID
	}
	return item
}

func toDomainUserTransitionConfirmation(entity *ent.UserTransitionConfirmation) *domain.UserTransitionConfirmation {
	item := &domain.UserTransitionConfirmation{
		ID:          entity.ID,
		Decision:    domain.UserTransitionDecision(entity.Decision),
		Comment:     entity.Comment,
		RespondedAt: entity.RespondedAt,
		CreatedAt:   entity.CreatedAt,
	}
	if entity.Edges.TransitionCase != nil {
		item.TransitionCaseID = entity.Edges.TransitionCase.ID
	}
	if entity.Edges.TransitionConfirmer != nil {
		item.TransitionConfirmerID = entity.Edges.TransitionConfirmer.ID
	}
	return item
}

func toDomainAccountHandlingTask(entity *ent.AccountHandlingTask) *domain.AccountHandlingTask {
	item := &domain.AccountHandlingTask{
		ID:                    entity.ID,
		ActionSnapshot:        domain.AccountActionField(entity.ActionSnapshot),
		ActionMessageSnapshot: entity.ActionMessageSnapshot,
		Status:                domain.AccountHandlingTaskStatus(entity.Status),
		StartedAt:             entity.StartedAt,
		CompletedAt:           entity.CompletedAt,
		FailureReason:         entity.FailureReason,
		RequestCount:          entity.RequestCount,
		CreatedAt:             entity.CreatedAt,
		UpdatedAt:             entity.UpdatedAt,
	}
	if entity.Edges.TransitionCase != nil {
		item.TransitionCaseID = entity.Edges.TransitionCase.ID
		if entity.Edges.TransitionCase.Edges.Owner != nil {
			item.UserID = entity.Edges.TransitionCase.Edges.Owner.ID
		}
	}
	if entity.Edges.Account != nil {
		item.AccountID = entity.Edges.Account.ID
	}
	if entity.Edges.Events != nil {
		item.Events = make([]domain.AccountHandlingEvent, len(entity.Edges.Events))
		for i, event := range entity.Edges.Events {
			eventItem := toDomainAccountHandlingEvent(event)
			eventItem.TaskID = entity.ID
			item.Events[i] = *eventItem
		}
	}
	if entity.Edges.Assignees != nil {
		item.Assignees = make([]domain.AccountHandlingTaskAssignee, len(entity.Edges.Assignees))
		for i, assignee := range entity.Edges.Assignees {
			assigneeItem := toDomainAccountHandlingTaskAssignee(assignee)
			assigneeItem.TaskID = entity.ID
			item.Assignees[i] = *assigneeItem
		}
	}
	return item
}

func toDomainAccountHandlingTaskAssignee(entity *ent.AccountHandlingTaskAssignee) *domain.AccountHandlingTaskAssignee {
	item := &domain.AccountHandlingTaskAssignee{
		ID:                    entity.ID,
		TaskID:                entity.TaskID,
		StakeholderID:         entity.StakeholderID,
		Name:                  entity.Name,
		Email:                 entity.Email,
		Phone:                 entity.Phone,
		Status:                domain.AccountHandlingTaskAssigneeStatus(entity.Status),
		RequestedAt:           entity.RequestedAt,
		RespondedAt:           entity.RespondedAt,
		ResponseMessage:       entity.ResponseMessage,
		RequestTokenExpiresAt: entity.RequestTokenExpiresAt,
		CreatedAt:             entity.CreatedAt,
		UpdatedAt:             entity.UpdatedAt,
	}
	if entity.RequestTokenHash != nil {
		item.RequestTokenHash = *entity.RequestTokenHash
	}
	if entity.Edges.Task != nil {
		item.TaskID = entity.Edges.Task.ID
	}
	return item
}

func toDomainAccountHandlingEvent(entity *ent.AccountHandlingEvent) *domain.AccountHandlingEvent {
	item := &domain.AccountHandlingEvent{
		ID:        entity.ID,
		EventType: domain.AccountHandlingEventType(entity.EventType),
		Message:   entity.Message,
		CreatedAt: entity.CreatedAt,
	}
	if entity.Edges.Task != nil {
		item.TaskID = entity.Edges.Task.ID
	}
	return item
}
