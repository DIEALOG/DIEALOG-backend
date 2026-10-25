package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent/enttest"
	"github.com/DIEALOG/DIEALOG-backend/infra/repository"
	"github.com/DIEALOG/DIEALOG-backend/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestTransitionFlow(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:transition_flow?mode=memory&_fk=1")
	defer client.Close()

	ctx := context.Background()
	timeout := 2 * time.Second

	userRepo := repository.NewUserRepository(client)
	legacyRepo := repository.NewLegacyRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)
	stakeholderGroupRepo := repository.NewStakeholderGroupRepository(client)
	policyRepo := repository.NewUserTransitionPolicyRepository(client)
	confirmerRepo := repository.NewUserTransitionConfirmerRepository(client)
	aliveRepo := repository.NewUserAliveCheckRepository(client)
	caseRepo := repository.NewUserTransitionCaseRepository(client)
	confirmationRepo := repository.NewUserTransitionConfirmationRepository(client)
	taskRepo := repository.NewAccountHandlingTaskRepository(client)
	eventRepo := repository.NewAccountHandlingEventRepository(client)

	policyUseCase := usecase.NewUserTransitionPolicyUseCase(policyRepo, repository.NewUserConfigRepository(client), timeout)
	confirmerUseCase := usecase.NewUserTransitionConfirmerUseCase(confirmerRepo, stakeholderRepo, timeout)
	aliveUseCase := usecase.NewUserAliveCheckUseCase(aliveRepo, policyUseCase, timeout)
	caseUseCase := usecase.NewUserTransitionCaseUseCase(caseRepo, confirmationRepo, policyUseCase, timeout)
	confirmationUseCase := usecase.NewUserTransitionConfirmationUseCase(confirmationRepo, caseRepo, caseUseCase, timeout)
	accountUseCase := usecase.NewAccountHandlingUseCase(taskRepo, eventRepo, legacyRepo, caseUseCase, timeout)

	owner, err := userRepo.Create(ctx, &domain.User{
		Name:        "owner",
		Email:       "owner@example.com",
		Password:    "password",
		PhoneNumber: "010-1234-5678",
	})
	require.NoError(t, err)

	t.Run("default policy and validation", func(t *testing.T) {
		policy, err := policyUseCase.Get(ctx, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, policy.MinimumConfirmers)
		assert.Equal(t, 7, policy.WaitingPeriodDays)
		assert.Equal(t, 90, policy.AliveCheckIntervalDays)

		_, err = policyUseCase.Update(ctx, &domain.UpdateUserTransitionPolicyRequest{
			MinimumConfirmers: intPointer(0),
		}, &owner.ID)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, 400, domainErr.StatusCode)
	})

	t.Run("invite, accept, and authenticate confirmer", func(t *testing.T) {
		email := "confirmer@example.com"
		invitation, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Confirmer",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)
		require.NotEmpty(t, invitation.InvitationToken)
		assert.Equal(t, domain.UserTransitionConfirmerStatusPending, invitation.Confirmer.Status)

		require.NotNil(t, invitation.Confirmer.Email)
		assert.Equal(t, "confirmer@example.com", *invitation.Confirmer.Email)

		_, err = confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Duplicate",
			Email: &email,
		}, &owner.ID)
		require.Error(t, err)

		mixedCaseEmail := "  CONFIRMER@EXAMPLE.COM "
		_, err = confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Duplicate mixed case",
			Email: &mixedCaseEmail,
		}, &owner.ID)
		require.Error(t, err)

		accepted, err := confirmerUseCase.Accept(ctx, invitation.InvitationToken)
		require.NoError(t, err)
		require.NotEmpty(t, accepted.AuthToken)
		assert.Equal(t, domain.UserTransitionConfirmerStatusAccepted, accepted.Confirmer.Status)

		authenticated, err := confirmerUseCase.Authenticate(ctx, accepted.AuthToken)
		require.NoError(t, err)
		assert.Equal(t, accepted.Confirmer.ID, authenticated.ID)

		_, err = confirmerUseCase.Accept(ctx, invitation.InvitationToken)
		require.Error(t, err)
	})

	t.Run("invitation resend and expiry", func(t *testing.T) {
		email := "resend@example.com"
		invitation, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Resend",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)

		resent, err := confirmerUseCase.Resend(ctx, &invitation.Confirmer.ID, &owner.ID)
		require.NoError(t, err)
		require.NotEmpty(t, resent.InvitationToken)

		_, err = confirmerUseCase.Accept(ctx, invitation.InvitationToken)
		require.Error(t, err)

		acceptedResent, err := confirmerUseCase.Accept(ctx, resent.InvitationToken)
		require.NoError(t, err)
		assert.Nil(t, acceptedResent.Confirmer.InviteExpiresAt)

		email = "expired@example.com"
		expiring, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Expired",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)

		stored, err := confirmerRepo.FindByID(ctx, &expiring.Confirmer.ID, &owner.ID)
		require.NoError(t, err)
		past := time.Now().Add(-time.Hour)
		stored.InviteExpiresAt = &past
		_, err = confirmerRepo.Update(ctx, stored, &owner.ID)
		require.NoError(t, err)

		_, err = confirmerUseCase.Accept(ctx, expiring.InvitationToken)
		require.Error(t, err)

		reinvited, err := confirmerUseCase.Resend(ctx, &expiring.Confirmer.ID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionConfirmerStatusPending, reinvited.Confirmer.Status)
		assert.Nil(t, reinvited.Confirmer.AcceptedAt)
		assert.Nil(t, reinvited.Confirmer.RevokedAt)
		assert.Nil(t, reinvited.Confirmer.AuthTokenExpiresAt)
		assert.NotNil(t, reinvited.Confirmer.InviteExpiresAt)
	})

	t.Run("transition flow with account handling", func(t *testing.T) {
		updatedPolicy, err := policyUseCase.Update(ctx, &domain.UpdateUserTransitionPolicyRequest{
			MinimumConfirmers:      intPointer(1),
			WaitingPeriodDays:      intPointer(0),
			AliveCheckIntervalDays: intPointer(1),
		}, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, updatedPolicy.MinimumConfirmers)

		email := "second-confirmer@example.com"
		invitation, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Second",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)
		accepted, err := confirmerUseCase.Accept(ctx, invitation.InvitationToken)
		require.NoError(t, err)

		handlerEmail := "handler@example.com"
		handler, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "Account Handler",
			Email: &handlerEmail,
		}, &owner.ID)
		require.NoError(t, err)
		groupHandlerEmail := "group-handler@example.com"
		groupHandler, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "Group Handler",
			Email: &groupHandlerEmail,
		}, &owner.ID)
		require.NoError(t, err)
		handlerGroup, err := stakeholderGroupRepo.Create(ctx, &domain.StakeholderGroup{
			Name:           "Account Handlers",
			StakeholderIDs: []domain.ID{handler.ID, groupHandler.ID},
		}, &owner.ID)
		require.NoError(t, err)

		legacy, err := legacyRepo.Create(ctx, &domain.Legacy{Name: "Social"}, &owner.ID)
		require.NoError(t, err)
		account, err := legacyRepo.CreateAccount(ctx, &domain.Account{
			Name: "Google",
			Type: domain.AccountType{Type: domain.AccountTypeGoogle},
			Action: domain.AccountAction{
				Action: domain.AccountActionRemain,
			},
			Stakeholders:      []domain.Stakeholder{*handler},
			StakeholderGroups: []domain.StakeholderGroup{*handlerGroup},
		}, &legacy.ID, &owner.ID)
		require.NoError(t, err)

		transitionCase, err := caseUseCase.Start(ctx, accepted.Confirmer)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionCaseStatusPending, transitionCase.Status)

		_, err = caseUseCase.Start(ctx, accepted.Confirmer)
		require.Error(t, err)

		_, err = confirmationUseCase.Submit(ctx, &transitionCase.ID, accepted.Confirmer, &domain.SubmitTransitionConfirmationRequest{
			Decision: domain.UserTransitionDecisionAgree,
		})
		require.NoError(t, err)

		confirmed, err := caseUseCase.GetByID(ctx, &transitionCase.ID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionCaseStatusConfirmed, confirmed.Status)
		require.NotNil(t, confirmed.ConfirmedAt)

		tasks, err := accountUseCase.ListByCaseID(ctx, &transitionCase.ID, &owner.ID)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, account.ID, tasks[0].AccountID)
		assert.Equal(t, domain.AccountActionRemain, tasks[0].ActionSnapshot)
		assert.Equal(t, domain.AccountHandlingTaskStatusPending, tasks[0].Status)
		require.Len(t, tasks[0].Assignees, 2)
		require.NotEmpty(t, tasks[0].Events)
		assert.Equal(t, tasks[0].ID, tasks[0].Events[0].TaskID)

		// The action is only a request to the stakeholders. It must not complete
		// without a stakeholder reporting success.
		assert.Equal(t, domain.AccountHandlingTaskStatusPending, tasks[0].Status)
		request, err := accountUseCase.Request(ctx, &tasks[0].ID, &owner.ID)
		require.NoError(t, err)
		require.NotNil(t, request.Task)
		assert.Equal(t, domain.AccountHandlingTaskStatusInProgress, request.Task.Status)
		assert.Equal(t, 1, request.Task.RequestCount)
		require.Len(t, request.Requests, 2)

		completed, err := accountUseCase.SubmitResult(ctx, &domain.SubmitAccountHandlingResultRequest{
			Token:   request.Requests[0].Token,
			Success: boolPointer(true),
			Message: stringPointer("account left as-is by the stakeholder"),
		})
		require.NoError(t, err)
		assert.Equal(t, domain.AccountHandlingTaskStatusCompleted, completed.Status)
		assert.Equal(t, 1, completed.RequestCount)
		assert.NotNil(t, completed.CompletedAt)
		require.Len(t, completed.Assignees, 2)
		var confirmedAssignees, cancelledAssignees int
		for _, assignee := range completed.Assignees {
			switch assignee.Status {
			case domain.AccountHandlingTaskAssigneeStatusConfirmed:
				confirmedAssignees++
			case domain.AccountHandlingTaskAssigneeStatusCancelled:
				cancelledAssignees++
			}
		}
		assert.Equal(t, 1, confirmedAssignees)
		assert.Equal(t, 1, cancelledAssignees)

		replayed, err := accountUseCase.SubmitResult(ctx, &domain.SubmitAccountHandlingResultRequest{
			Token:   request.Requests[0].Token,
			Success: boolPointer(true),
		})
		require.NoError(t, err)
		assert.Equal(t, domain.AccountHandlingTaskStatusCompleted, replayed.Status)
		assert.Len(t, replayed.Events, len(completed.Events))
		assert.Equal(t, 1, confirmedAssignees)

		_, err = accountUseCase.SubmitResult(ctx, &domain.SubmitAccountHandlingResultRequest{
			Token:   request.Requests[0].Token,
			Success: boolPointer(false),
		})
		require.Error(t, err)

		idempotent, err := accountUseCase.Request(ctx, &tasks[0].ID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.AccountHandlingTaskStatusCompleted, idempotent.Task.Status)
		assert.Empty(t, idempotent.Requests)
		assert.Len(t, idempotent.Task.Events, len(completed.Events))

		deleteAccount, err := legacyRepo.CreateAccount(ctx, &domain.Account{
			Name: "Delete me",
			Type: domain.AccountType{Type: domain.AccountTypeGoogle},
			Action: domain.AccountAction{
				Action: domain.AccountActionDelete,
			},
			Stakeholders: []domain.Stakeholder{*handler},
		}, &legacy.ID, &owner.ID)
		require.NoError(t, err)
		noHandlerAccount, err := legacyRepo.CreateAccount(ctx, &domain.Account{
			Name: "No handler",
			Type: domain.AccountType{Type: domain.AccountTypeGoogle},
			Action: domain.AccountAction{
				Action: domain.AccountActionRemain,
			},
		}, &legacy.ID, &owner.ID)
		require.NoError(t, err)

		secondCase, err := caseUseCase.Start(ctx, accepted.Confirmer)
		require.NoError(t, err)
		_, err = confirmationUseCase.Submit(ctx, &secondCase.ID, accepted.Confirmer, &domain.SubmitTransitionConfirmationRequest{
			Decision: domain.UserTransitionDecisionAgree,
		})
		require.NoError(t, err)

		secondTasks, err := accountUseCase.ListByCaseID(ctx, &secondCase.ID, &owner.ID)
		require.NoError(t, err)
		require.Len(t, secondTasks, 3)

		var deleteTaskID, noHandlerTaskID domain.ID
		for _, task := range secondTasks {
			if task.AccountID == deleteAccount.ID {
				deleteTaskID = task.ID
			}
			if task.AccountID == noHandlerAccount.ID {
				noHandlerTaskID = task.ID
			}
		}
		require.NotEqual(t, domain.ID{}, deleteTaskID)
		require.NotEqual(t, domain.ID{}, noHandlerTaskID)

		_, err = accountUseCase.Request(ctx, &noHandlerTaskID, &owner.ID)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, 400, domainErr.StatusCode)

		// If a stakeholder is assigned after task creation, the request endpoint
		// snapshots the current account stakeholders before issuing tokens.
		_, err = client.Account.UpdateOneID(noHandlerAccount.ID).
			AddStakeholderIDs(handler.ID).
			Save(ctx)
		require.NoError(t, err)
		lateRequest, err := accountUseCase.Request(ctx, &noHandlerTaskID, &owner.ID)
		require.NoError(t, err)
		require.Len(t, lateRequest.Requests, 1)
		assert.Equal(t, handler.ID, lateRequest.Requests[0].StakeholderID)

		deleteRequest, err := accountUseCase.Request(ctx, &deleteTaskID, &owner.ID)
		require.NoError(t, err)
		require.Len(t, deleteRequest.Requests, 1)
		assert.Equal(t, 1, deleteRequest.Task.RequestCount)

		failed, err := accountUseCase.SubmitResult(ctx, &domain.SubmitAccountHandlingResultRequest{
			Token:   deleteRequest.Requests[0].Token,
			Success: boolPointer(false),
			Message: stringPointer("account deletion failed"),
		})
		require.NoError(t, err)
		assert.Equal(t, domain.AccountHandlingTaskStatusFailed, failed.Status)
		assert.Equal(t, 1, failed.RequestCount)
		require.NotNil(t, failed.FailureReason)
		assert.Equal(t, "account deletion failed", *failed.FailureReason)

		// A request refreshes the assignee snapshot additively, so a stakeholder
		// added after the task was created receives a token on the next request.
		_, err = client.Account.UpdateOneID(deleteAccount.ID).
			AddStakeholderIDs(groupHandler.ID).
			Save(ctx)
		require.NoError(t, err)
		retriedRequest, err := accountUseCase.Request(ctx, &deleteTaskID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.AccountHandlingTaskStatusInProgress, retriedRequest.Task.Status)
		assert.Equal(t, 2, retriedRequest.Task.RequestCount)
		require.Len(t, retriedRequest.Requests, 2)
		require.Len(t, retriedRequest.Task.Assignees, 2)

		recovered, err := accountUseCase.SubmitResult(ctx, &domain.SubmitAccountHandlingResultRequest{
			Token:   retriedRequest.Requests[0].Token,
			Success: boolPointer(true),
		})
		require.NoError(t, err)
		assert.Equal(t, domain.AccountHandlingTaskStatusCompleted, recovered.Status)
		assert.Equal(t, 2, recovered.RequestCount)
		assert.Nil(t, recovered.FailureReason)
		assert.NotNil(t, recovered.CompletedAt)
	})

	t.Run("waiting period and cancellation", func(t *testing.T) {
		_, err := policyUseCase.Update(ctx, &domain.UpdateUserTransitionPolicyRequest{
			MinimumConfirmers:      intPointer(1),
			WaitingPeriodDays:      intPointer(1),
			AliveCheckIntervalDays: intPointer(1),
		}, &owner.ID)
		require.NoError(t, err)

		email := "third-confirmer@example.com"
		invitation, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "Third",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)
		accepted, err := confirmerUseCase.Accept(ctx, invitation.InvitationToken)
		require.NoError(t, err)

		transitionCase, err := caseUseCase.Start(ctx, accepted.Confirmer)
		require.NoError(t, err)

		_, err = confirmationUseCase.Submit(ctx, &transitionCase.ID, accepted.Confirmer, &domain.SubmitTransitionConfirmationRequest{
			Decision: domain.UserTransitionDecisionAgree,
		})
		require.NoError(t, err)

		waiting, err := caseUseCase.GetByID(ctx, &transitionCase.ID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionCaseStatusWaiting, waiting.Status)
		require.NotNil(t, waiting.EligibleAt)
		assert.True(t, waiting.EligibleAt.After(time.Now()))

		cancelled, err := caseUseCase.Cancel(ctx, &transitionCase.ID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionCaseStatusCancelled, cancelled.Status)
		require.NotNil(t, cancelled.CancelledAt)

		restarted, err := caseUseCase.Start(ctx, accepted.Confirmer)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionCaseStatusPending, restarted.Status)
		_, err = caseUseCase.Cancel(ctx, &restarted.ID, &owner.ID)
		require.NoError(t, err)
	})

	t.Run("alive check scheduling and response", func(t *testing.T) {
		checks, err := aliveUseCase.ListByUserID(ctx, &owner.ID)
		require.NoError(t, err)
		require.Len(t, checks, 1)
		first := checks[0]
		assert.Equal(t, domain.UserAliveCheckStatusScheduled, first.Status)

		again, err := aliveUseCase.ScheduleNext(ctx, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, first.ID, again.ID)

		responded, err := aliveUseCase.Respond(ctx, &first.ID, &owner.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.UserAliveCheckStatusResponded, responded.Status)

		checks, err = aliveUseCase.ListByUserID(ctx, &owner.ID)
		require.NoError(t, err)
		require.Len(t, checks, 2)
		assert.Equal(t, 2, checks[0].Sequence)
	})
}

func intPointer(value int) *int {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}

func stringPointer(value string) *string {
	return &value
}
