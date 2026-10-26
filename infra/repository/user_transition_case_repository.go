package repository

import (
	"context"
	"sort"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/account"
	"github.com/DIEALOG/DIEALOG-backend/ent/accounthandlingtask"
	"github.com/DIEALOG/DIEALOG-backend/ent/legacy"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitioncase"
)

type userTransitionCaseRepository struct {
	client *ent.Client
}

func NewUserTransitionCaseRepository(client *ent.Client) domain.UserTransitionCaseRepository {
	return &userTransitionCaseRepository{client: client}
}

func (r *userTransitionCaseRepository) Create(c context.Context, transitionCase *domain.UserTransitionCase, userID *domain.ID) (*domain.UserTransitionCase, error) {
	tx, err := r.client.Tx(c)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	active, err := tx.UserTransitionCase.Query().
		Where(
			usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			usertransitioncase.StatusIn(
				string(domain.UserTransitionCaseStatusPending),
				string(domain.UserTransitionCaseStatusWaiting),
			),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, domain.ErrActiveTransitionCaseExists
	}

	entity, err := tx.UserTransitionCase.Create().
		SetActiveSlot(userID.String()).
		SetStatus(string(transitionCase.Status)).
		SetMinimumConfirmersSnapshot(transitionCase.MinimumConfirmersSnapshot).
		SetWaitingPeriodDaysSnapshot(transitionCase.WaitingPeriodDaysSnapshot).
		SetRequestedAt(transitionCase.RequestedAt).
		SetOwnerID(*userID).
		Save(c)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.ErrActiveTransitionCaseExists
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	item := toDomainUserTransitionCase(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userTransitionCaseRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserTransitionCase, error) {
	entity, err := r.client.UserTransitionCase.Query().
		Where(
			usertransitioncase.IDEQ(*id),
			usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionCase(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userTransitionCaseRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.UserTransitionCase, error) {
	entities, err := r.client.UserTransitionCase.Query().
		Where(usertransitioncase.HasOwnerWith(user.IDEQ(*userID))).
		WithOwner().
		Order(ent.Desc(usertransitioncase.FieldRequestedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.UserTransitionCase, len(entities))
	for i, entity := range entities {
		items[i] = toDomainUserTransitionCase(entity)
		items[i].UserID = *userID
	}
	return items, nil
}

func (r *userTransitionCaseRepository) FindActiveByUserID(c context.Context, userID *domain.ID) (*domain.UserTransitionCase, error) {
	entity, err := r.client.UserTransitionCase.Query().
		Where(
			usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			usertransitioncase.StatusIn(
				string(domain.UserTransitionCaseStatusPending),
				string(domain.UserTransitionCaseStatusWaiting),
			),
		).
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionCase(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userTransitionCaseRepository) Update(c context.Context, transitionCase *domain.UserTransitionCase, userID *domain.ID) (*domain.UserTransitionCase, error) {
	builder := r.client.UserTransitionCase.Update().
		Where(
			usertransitioncase.IDEQ(transitionCase.ID),
			usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
		)

	// Guard the state machine at the database level so a concurrent cancel or
	// confirmation cannot be overwritten by a stale read.
	switch transitionCase.Status {
	case domain.UserTransitionCaseStatusWaiting:
		builder.Where(usertransitioncase.StatusEQ(string(domain.UserTransitionCaseStatusPending)))
	case domain.UserTransitionCaseStatusCancelled:
		builder.Where(usertransitioncase.StatusIn(
			string(domain.UserTransitionCaseStatusPending),
			string(domain.UserTransitionCaseStatusWaiting),
		))
	case domain.UserTransitionCaseStatusConfirmed:
		builder.Where(usertransitioncase.StatusEQ(string(domain.UserTransitionCaseStatusWaiting)))
	case domain.UserTransitionCaseStatusPending:
		builder.Where(usertransitioncase.StatusEQ(string(domain.UserTransitionCaseStatusPending)))
	}

	builder.
		SetStatus(string(transitionCase.Status)).
		SetMinimumConfirmersSnapshot(transitionCase.MinimumConfirmersSnapshot).
		SetWaitingPeriodDaysSnapshot(transitionCase.WaitingPeriodDaysSnapshot).
		SetRequestedAt(transitionCase.RequestedAt)

	if transitionCase.Status == domain.UserTransitionCaseStatusConfirmed ||
		transitionCase.Status == domain.UserTransitionCaseStatusCancelled {
		builder.ClearActiveSlot()
	} else {
		builder.SetActiveSlot(userID.String())
	}
	if transitionCase.WaitingStartedAt != nil {
		builder.SetWaitingStartedAt(*transitionCase.WaitingStartedAt)
	} else {
		builder.ClearWaitingStartedAt()
	}
	if transitionCase.EligibleAt != nil {
		builder.SetEligibleAt(*transitionCase.EligibleAt)
	} else {
		builder.ClearEligibleAt()
	}
	if transitionCase.ConfirmedAt != nil {
		builder.SetConfirmedAt(*transitionCase.ConfirmedAt)
	} else {
		builder.ClearConfirmedAt()
	}
	if transitionCase.CancelledAt != nil {
		builder.SetCancelledAt(*transitionCase.CancelledAt)
	} else {
		builder.ClearCancelledAt()
	}

	n, err := builder.Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		current, findErr := r.FindByID(c, &transitionCase.ID, userID)
		if findErr != nil {
			return nil, findErr
		}
		if current.Status == transitionCase.Status {
			return current, nil
		}
		if transitionCase.Status == domain.UserTransitionCaseStatusWaiting &&
			(current.Status == domain.UserTransitionCaseStatusCancelled ||
				current.Status == domain.UserTransitionCaseStatusConfirmed) {
			return current, nil
		}
		return nil, domain.ErrInvalidTransitionState
	}

	return r.FindByID(c, &transitionCase.ID, userID)
}

func (r *userTransitionCaseRepository) ConfirmAndCreateTasks(c context.Context, id *domain.ID, userID *domain.ID, now time.Time) (*domain.UserTransitionCase, error) {
	tx, err := r.client.Tx(c)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	transitionCase, err := tx.UserTransitionCase.Query().
		Where(
			usertransitioncase.IDEQ(*id),
			usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}
	if transitionCase.Status != string(domain.UserTransitionCaseStatusWaiting) ||
		transitionCase.EligibleAt == nil ||
		now.Before(*transitionCase.EligibleAt) {
		return nil, domain.ErrTransitionCaseNotEligible
	}

	accounts, err := tx.Account.Query().
		Where(account.HasLegacyWith(legacy.HasOwnerWith(user.IDEQ(*userID)))).
		WithStakeholders().
		WithStakeholderGroups(func(q *ent.StakeholderGroupQuery) {
			q.WithStakeholders()
		}).
		All(c)
	if err != nil {
		return nil, err
	}

	n, err := tx.UserTransitionCase.Update().
		Where(
			usertransitioncase.IDEQ(*id),
			usertransitioncase.StatusEQ(string(domain.UserTransitionCaseStatusWaiting)),
		).
		SetStatus(string(domain.UserTransitionCaseStatusConfirmed)).
		SetConfirmedAt(now).
		ClearActiveSlot().
		SetUpdatedAt(now).
		Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, domain.ErrTransitionCaseNotEligible
	}

	for _, item := range accounts {
		exists, err := tx.AccountHandlingTask.Query().
			Where(
				accounthandlingtask.HasTransitionCaseWith(usertransitioncase.IDEQ(*id)),
				accounthandlingtask.HasAccountWith(account.IDEQ(item.ID)),
			).
			Exist(c)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		task, err := tx.AccountHandlingTask.Create().
			SetTransitionCaseID(*id).
			SetAccountID(item.ID).
			SetActionSnapshot(string(item.Action)).
			SetNillableActionMessageSnapshot(item.ActionMessage).
			SetStatus(string(domain.AccountHandlingTaskStatusPending)).
			Save(c)
		if err != nil {
			return nil, err
		}

		assignees := accountHandlingAssignees(item)
		message := "account handling task created"
		if len(assignees) == 0 {
			message = "account handling task created; no stakeholders assigned"
		}
		_, err = tx.AccountHandlingEvent.Create().
			SetTaskID(task.ID).
			SetEventType(string(domain.AccountHandlingEventTypeTaskCreated)).
			SetMessage(message).
			Save(c)
		if err != nil {
			return nil, err
		}

		for _, assignee := range assignees {
			if _, err := tx.AccountHandlingTaskAssignee.Create().
				SetTaskID(task.ID).
				SetStakeholderID(assignee.StakeholderID).
				SetName(assignee.Name).
				SetNillableEmail(assignee.Email).
				SetNillablePhone(assignee.Phone).
				SetStatus(string(domain.AccountHandlingTaskAssigneeStatusPending)).
				Save(c); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(c, id, userID)
}

func accountHandlingAssignees(item *ent.Account) []domain.AccountHandlingTaskAssignee {
	seen := make(map[domain.ID]struct{})
	assignees := make([]domain.AccountHandlingTaskAssignee, 0)
	add := func(stakeholder *ent.Stakeholder) {
		if stakeholder == nil {
			return
		}
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

	for _, stakeholder := range item.Edges.Stakeholders {
		add(stakeholder)
	}
	for _, group := range item.Edges.StakeholderGroups {
		for _, stakeholder := range group.Edges.Stakeholders {
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
