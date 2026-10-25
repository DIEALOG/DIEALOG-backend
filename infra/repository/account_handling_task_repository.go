package repository

import (
	"context"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/accounthandlingevent"
	"github.com/DIEALOG/DIEALOG-backend/ent/accounthandlingtask"
	"github.com/DIEALOG/DIEALOG-backend/ent/accounthandlingtaskassignee"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitioncase"
)

type accountHandlingTaskRepository struct {
	client *ent.Client
}

func NewAccountHandlingTaskRepository(client *ent.Client) domain.AccountHandlingTaskRepository {
	return &accountHandlingTaskRepository{client: client}
}

func (r *accountHandlingTaskRepository) FindAllByCaseID(c context.Context, caseID *domain.ID, userID *domain.ID) ([]*domain.AccountHandlingTask, error) {
	entities, err := r.client.AccountHandlingTask.Query().
		Where(
			accounthandlingtask.HasTransitionCaseWith(
				usertransitioncase.IDEQ(*caseID),
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
		).
		WithTransitionCase(func(q *ent.UserTransitionCaseQuery) {
			q.WithOwner()
		}).
		WithAccount().
		WithAssignees(func(q *ent.AccountHandlingTaskAssigneeQuery) {
			q.Order(ent.Asc(accounthandlingtaskassignee.FieldName), ent.Asc(accounthandlingtaskassignee.FieldID))
		}).
		WithEvents(func(q *ent.AccountHandlingEventQuery) {
			q.Order(ent.Asc(accounthandlingevent.FieldCreatedAt))
		}).
		Order(ent.Asc(accounthandlingtask.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.AccountHandlingTask, len(entities))
	for i, entity := range entities {
		items[i] = toDomainAccountHandlingTask(entity)
	}
	return items, nil
}

func (r *accountHandlingTaskRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.AccountHandlingTask, error) {
	entity, err := r.client.AccountHandlingTask.Query().
		Where(
			accounthandlingtask.IDEQ(*id),
			accounthandlingtask.HasTransitionCaseWith(
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
		).
		WithTransitionCase(func(q *ent.UserTransitionCaseQuery) {
			q.WithOwner()
		}).
		WithAccount().
		WithAssignees(func(q *ent.AccountHandlingTaskAssigneeQuery) {
			q.Order(ent.Asc(accounthandlingtaskassignee.FieldName), ent.Asc(accounthandlingtaskassignee.FieldID))
		}).
		WithEvents(func(q *ent.AccountHandlingEventQuery) {
			q.Order(ent.Asc(accounthandlingevent.FieldCreatedAt))
		}).
		Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainAccountHandlingTask(entity), nil
}

func (r *accountHandlingTaskRepository) FindByAssigneeTokenHash(c context.Context, tokenHash string) (*domain.AccountHandlingTask, *domain.AccountHandlingTaskAssignee, error) {
	entity, err := r.client.AccountHandlingTaskAssignee.Query().
		Where(accounthandlingtaskassignee.RequestTokenHashEQ(tokenHash)).
		WithTask(func(q *ent.AccountHandlingTaskQuery) {
			q.WithTransitionCase(func(q *ent.UserTransitionCaseQuery) {
				q.WithOwner()
			}).
				WithAccount().
				WithAssignees(func(q *ent.AccountHandlingTaskAssigneeQuery) {
					q.Order(ent.Asc(accounthandlingtaskassignee.FieldName), ent.Asc(accounthandlingtaskassignee.FieldID))
				}).
				WithEvents(func(q *ent.AccountHandlingEventQuery) {
					q.Order(ent.Asc(accounthandlingevent.FieldCreatedAt))
				})
		}).
		Only(c)
	if err != nil {
		return nil, nil, err
	}
	if entity.Edges.Task == nil {
		return nil, nil, &ent.NotFoundError{}
	}

	return toDomainAccountHandlingTask(entity.Edges.Task), toDomainAccountHandlingTaskAssignee(entity), nil
}

func (r *accountHandlingTaskRepository) EnsureAssignees(c context.Context, taskID *domain.ID, userID *domain.ID, assignees []domain.AccountHandlingTaskAssignee) (*domain.AccountHandlingTask, error) {
	tx, err := r.client.Tx(c)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	exists, err := tx.AccountHandlingTask.Query().
		Where(
			accounthandlingtask.IDEQ(*taskID),
			accounthandlingtask.HasTransitionCaseWith(
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	for _, assignee := range assignees {
		if assignee.StakeholderID == (domain.ID{}) {
			continue
		}
		alreadyExists, err := tx.AccountHandlingTaskAssignee.Query().
			Where(
				accounthandlingtaskassignee.HasTaskWith(accounthandlingtask.IDEQ(*taskID)),
				accounthandlingtaskassignee.StakeholderIDEQ(assignee.StakeholderID),
			).
			Exist(c)
		if err != nil {
			return nil, err
		}
		if alreadyExists {
			continue
		}

		if _, err := tx.AccountHandlingTaskAssignee.Create().
			SetTaskID(*taskID).
			SetStakeholderID(assignee.StakeholderID).
			SetName(assignee.Name).
			SetNillableEmail(assignee.Email).
			SetNillablePhone(assignee.Phone).
			SetStatus(string(domain.AccountHandlingTaskAssigneeStatusPending)).
			Save(c); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindByID(c, taskID, userID)
}

func (r *accountHandlingTaskRepository) SaveRequest(c context.Context, task *domain.AccountHandlingTask, assignees []domain.AccountHandlingTaskAssignee, userID *domain.ID) (*domain.AccountHandlingTask, error) {
	tx, err := r.client.Tx(c)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	builder := tx.AccountHandlingTask.Update().
		Where(
			accounthandlingtask.IDEQ(task.ID),
			accounthandlingtask.HasTransitionCaseWith(
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
			accounthandlingtask.StatusIn(
				string(domain.AccountHandlingTaskStatusPending),
				string(domain.AccountHandlingTaskStatusInProgress),
				string(domain.AccountHandlingTaskStatusFailed),
			),
		).
		SetStatus(string(domain.AccountHandlingTaskStatusInProgress)).
		AddRequestCount(1).
		SetUpdatedAt(now).
		ClearCompletedAt().
		ClearFailureReason()
	if task.StartedAt != nil {
		builder.SetStartedAt(*task.StartedAt)
	}
	n, err := builder.Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		_ = tx.Rollback()
		current, findErr := r.FindByID(c, &task.ID, userID)
		if findErr == nil && current.Status == domain.AccountHandlingTaskStatusCompleted {
			return current, nil
		}
		if findErr != nil {
			return nil, findErr
		}
		return nil, domain.ErrInvalidTransitionState
	}

	for _, assignee := range assignees {
		builder := tx.AccountHandlingTaskAssignee.Update().
			Where(
				accounthandlingtaskassignee.IDEQ(assignee.ID),
				accounthandlingtaskassignee.HasTaskWith(accounthandlingtask.IDEQ(task.ID)),
				accounthandlingtaskassignee.StatusIn(
					string(domain.AccountHandlingTaskAssigneeStatusPending),
					string(domain.AccountHandlingTaskAssigneeStatusFailed),
					string(domain.AccountHandlingTaskAssigneeStatusCancelled),
				),
			).
			SetStatus(string(domain.AccountHandlingTaskAssigneeStatusPending)).
			SetUpdatedAt(now).
			ClearRespondedAt().
			ClearResponseMessage()
		if assignee.RequestTokenHash != "" {
			builder.SetRequestTokenHash(assignee.RequestTokenHash)
		} else {
			builder.ClearRequestTokenHash()
		}
		if assignee.RequestTokenExpiresAt != nil {
			builder.SetRequestTokenExpiresAt(*assignee.RequestTokenExpiresAt)
		} else {
			builder.ClearRequestTokenExpiresAt()
		}
		if assignee.RequestedAt != nil {
			builder.SetRequestedAt(*assignee.RequestedAt)
		} else {
			builder.ClearRequestedAt()
		}
		if _, err := builder.Save(c); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindByID(c, &task.ID, userID)
}

func (r *accountHandlingTaskRepository) SaveResult(c context.Context, task *domain.AccountHandlingTask, assignee *domain.AccountHandlingTaskAssignee, userID *domain.ID) (*domain.AccountHandlingTask, error) {
	tx, err := r.client.Tx(c)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	builder := tx.AccountHandlingTask.Update().
		Where(
			accounthandlingtask.IDEQ(task.ID),
			accounthandlingtask.HasTransitionCaseWith(
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
			accounthandlingtask.StatusEQ(string(domain.AccountHandlingTaskStatusInProgress)),
		).
		SetStatus(string(task.Status)).
		SetUpdatedAt(now)
	if task.CompletedAt != nil {
		builder.SetCompletedAt(*task.CompletedAt)
	} else {
		builder.ClearCompletedAt()
	}
	if task.FailureReason != nil {
		builder.SetFailureReason(*task.FailureReason)
	} else {
		builder.ClearFailureReason()
	}
	n, err := builder.Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, domain.ErrInvalidTransitionState
	}

	assigneeBuilder := tx.AccountHandlingTaskAssignee.Update().
		Where(
			accounthandlingtaskassignee.IDEQ(assignee.ID),
			accounthandlingtaskassignee.HasTaskWith(accounthandlingtask.IDEQ(task.ID)),
			accounthandlingtaskassignee.StatusEQ(string(domain.AccountHandlingTaskAssigneeStatusPending)),
			accounthandlingtaskassignee.RequestTokenHashEQ(assignee.RequestTokenHash),
		).
		SetStatus(string(assignee.Status)).
		SetUpdatedAt(now)
	// Keep the used token hash so a retried result submission can be
	// recognized as an idempotent replay. The response fields are set below.
	if assignee.RespondedAt != nil {
		assigneeBuilder.SetRespondedAt(*assignee.RespondedAt)
	} else {
		assigneeBuilder.ClearRespondedAt()
	}
	if assignee.ResponseMessage != nil {
		assigneeBuilder.SetResponseMessage(*assignee.ResponseMessage)
	} else {
		assigneeBuilder.ClearResponseMessage()
	}
	n, err = assigneeBuilder.Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, domain.ErrInvalidTransitionState
	}

	// Only one stakeholder response is needed. Once a result is recorded, the
	// remaining requests are no longer valid.
	if _, err := tx.AccountHandlingTaskAssignee.Update().
		Where(
			accounthandlingtaskassignee.HasTaskWith(accounthandlingtask.IDEQ(task.ID)),
			accounthandlingtaskassignee.IDNEQ(assignee.ID),
			accounthandlingtaskassignee.StatusEQ(string(domain.AccountHandlingTaskAssigneeStatusPending)),
		).
		SetStatus(string(domain.AccountHandlingTaskAssigneeStatusCancelled)).
		SetUpdatedAt(now).
		ClearRequestTokenHash().
		ClearRequestTokenExpiresAt().
		Save(c); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindByID(c, &task.ID, userID)
}
