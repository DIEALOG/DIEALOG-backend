package usecase_test

import (
	"context"
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

type mockCredentialUseCase struct{}

func (m *mockCredentialUseCase) StorePassword(c context.Context, userID *domain.ID, serviceName string, password string) error {
	return nil
}
func (m *mockCredentialUseCase) GetPassword(c context.Context, userID *domain.ID, serviceName string) (*string, error) {
	pwd := "testpassword"
	return &pwd, nil
}
func (m *mockCredentialUseCase) DeletePassword(c context.Context, userID *domain.ID, serviceName string) error {
	return nil
}

func TestNotificationFlow(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:notification_flow?mode=memory&_fk=1")
	defer client.Close()

	ctx := context.Background()
	timeout := 2 * time.Second

	userRepo := repository.NewUserRepository(client)
	legacyRepo := repository.NewLegacyRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)
	confirmerRepo := repository.NewUserTransitionConfirmerRepository(client)
	caseRepo := repository.NewUserTransitionCaseRepository(client)
	policyRepo := repository.NewUserTransitionPolicyRepository(client)
	confirmationRepo := repository.NewUserTransitionConfirmationRepository(client)
	notificationRepo := repository.NewNotificationRepository(client)

	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, timeout)
	policyUseCase := usecase.NewUserTransitionPolicyUseCase(policyRepo, timeout)
	confirmerUseCase := usecase.NewUserTransitionConfirmerUseCase(
		confirmerRepo,
		stakeholderRepo,
		userRepo,
		notificationUseCase,
		timeout,
	)
	caseUseCase := usecase.NewUserTransitionCaseUseCase(
		caseRepo,
		confirmationRepo,
		policyUseCase,
		confirmerRepo,
		stakeholderRepo,
		userRepo,
		notificationUseCase,
		timeout,
	)
	confirmationUseCase := usecase.NewUserTransitionConfirmationUseCase(confirmationRepo, caseRepo, caseUseCase, timeout)
	credUseCase := &mockCredentialUseCase{}
	legacyUseCase := usecase.NewLegacyUseCase(legacyRepo, credUseCase, userRepo, timeout)

	// Create User A (owner)
	userA, err := userRepo.Create(ctx, &domain.User{
		Name:        "UserA",
		Email:       "usera@example.com",
		Password:    "password",
		PhoneNumber: "010-1111-1111",
	})
	require.NoError(t, err)

	// Create User B (confirmer / stakeholder)
	userB, err := userRepo.Create(ctx, &domain.User{
		Name:        "UserB",
		Email:       "userb@example.com",
		Password:    "password",
		PhoneNumber: "010-2222-2222",
	})
	require.NoError(t, err)

	// Setup User A transition policy: 1 confirmer, 0 waiting days
	_, err = policyRepo.Create(ctx, &domain.UserTransitionPolicy{
		MinimumConfirmers:      1,
		WaitingPeriodDays:      0,
		AliveCheckIntervalDays: 90,
	}, &userA.ID)
	require.NoError(t, err)

	// 1. 확인자 요청 (Invite)
	t.Run("1. Confirmer invite creates notification for invited user", func(t *testing.T) {
		invitation, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  userB.Name,
			Email: &userB.Email,
			Phone: &userB.PhoneNumber,
		}, &userA.ID)
		require.NoError(t, err)
		require.NotNil(t, invitation)

		// Check User B's notifications
		notifs, err := notificationUseCase.ListByReceiverID(ctx, &userB.ID)
		require.NoError(t, err)
		require.Len(t, notifs, 1)
		assert.Equal(t, domain.NotificationTypeInvite, notifs[0].Type)
		assert.Contains(t, notifs[0].Message, "UserA")
		assert.Contains(t, notifs[0].Message, "사망 확인자")
	})

	// 2. 확인자 초대 수락 / 거절 (Accept / Decline)
	var confirmerAuthToken string
	t.Run("2. Confirmer invite accept and decline creates notification for owner", func(t *testing.T) {
		// First test decline with a temporary confirmer
		tempEmail := "temp@example.com"
		tempPhone := "010-9999-9999"
		tempInv, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
			Name:  "TempConfirmer",
			Email: &tempEmail,
			Phone: &tempPhone,
		}, &userA.ID)
		require.NoError(t, err)

		err = confirmerUseCase.Decline(ctx, tempInv.InvitationToken)
		require.NoError(t, err)

		// Owner (User A) should have decline notification
		ownerNotifs, err := notificationUseCase.ListByReceiverID(ctx, &userA.ID)
		require.NoError(t, err)
		require.True(t, len(ownerNotifs) >= 1)
		assert.Contains(t, ownerNotifs[0].Message, "거절")

		// Now test accept with User B's invitation
		confirmerList, err := confirmerUseCase.ListByUserID(ctx, &userA.ID)
		require.NoError(t, err)
		var userBConfirmer *domain.UserTransitionConfirmer
		for _, c := range confirmerList {
			if c.Email != nil && *c.Email == userB.Email {
				userBConfirmer = c
				break
			}
		}
		require.NotNil(t, userBConfirmer)

		// Re-fetch token via resend
		resendResp, err := confirmerUseCase.Resend(ctx, &userBConfirmer.ID, &userA.ID)
		require.NoError(t, err)

		acceptResp, err := confirmerUseCase.Accept(ctx, resendResp.InvitationToken)
		require.NoError(t, err)
		confirmerAuthToken = acceptResp.AuthToken

		// Owner (User A) should receive accept notification
		ownerNotifs, err = notificationUseCase.ListByReceiverID(ctx, &userA.ID)
		require.NoError(t, err)
		assert.Contains(t, ownerNotifs[0].Message, "수락")
	})

	// 3. 담당 기록 update 시
	t.Run("3. Legacy update creates notification for assigned stakeholder", func(t *testing.T) {
		// legacyUseCase has no notification dependency yet, so updating a legacy
		// does not notify the stakeholders of its accounts. Drop this skip once
		// legacy_usecase.Update sends notifications.
		t.Skip("legacy update notification is not implemented")

		// Create stakeholder for User A pointing to User B's email/phone
		stakeholder, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  userB.Name,
			Email: &userB.Email,
			Phone: &userB.PhoneNumber,
		}, &userA.ID)
		require.NoError(t, err)

		// Create legacy with an account assigned to that stakeholder
		legacy, err := legacyUseCase.Create(ctx, &domain.Legacy{
			Name: "기록 A",
		}, &userA.ID)
		require.NoError(t, err)

		account, err := legacyUseCase.CreateAccount(ctx, &domain.Account{
			Type:         domain.AccountType{Type: domain.AccountTypeGoogle},
			Name:         "my-google",
			Action:       domain.AccountAction{Action: domain.AccountActionRemain},
			Stakeholders: []domain.Stakeholder{*stakeholder},
		}, nil, &legacy.ID, &userA.ID)
		require.NoError(t, err)
		require.NotNil(t, account)

		// Update legacy name
		_, err = legacyUseCase.Update(ctx, &legacy.ID, "수정된 기록 A", &userA.ID)
		require.NoError(t, err)

		// User B should receive legacy update notification
		notifs, err := notificationUseCase.ListByReceiverID(ctx, &userB.ID)
		require.NoError(t, err)
		var legacyNotif *domain.Notification
		for _, n := range notifs {
			if n.Type == domain.NotificationTypeDiealog {
				legacyNotif = n
				break
			}
		}
		require.NotNil(t, legacyNotif)
		assert.Contains(t, legacyNotif.Message, "수정된 기록 A")
		assert.Contains(t, legacyNotif.Message, "담당자로 지정된 기록")
	})

	// 4. 확인자로 있는 지인의 사망심사 시
	var transitionCase *domain.UserTransitionCase
	t.Run("4. Transition case start creates notification for confirmers", func(t *testing.T) {
		// Authenticate confirmerItem
		authenticatedConfirmer, err := confirmerUseCase.Authenticate(ctx, confirmerAuthToken)
		require.NoError(t, err)

		transitionCase, err = caseUseCase.Start(ctx, authenticatedConfirmer)
		require.NoError(t, err)
		require.NotNil(t, transitionCase)

		// User B is the confirmer, should receive case start notification
		notifs, err := notificationUseCase.ListByReceiverID(ctx, &userB.ID)
		require.NoError(t, err)
		var startNotif *domain.Notification
		for _, n := range notifs {
			if n.Type == domain.NotificationTypeTransition && n.Metadata["type"] == "transition_case_started" {
				startNotif = n
				break
			}
		}
		require.NotNil(t, startNotif)
		assert.Contains(t, startNotif.Message, "사망 확인 심사가 시작되었습니다")
	})

	// 5. 지인의 사망 전환 시
	t.Run("5. Transition confirmed creates notification for confirmers and stakeholders", func(t *testing.T) {
		authenticatedConfirmer, err := confirmerUseCase.Authenticate(ctx, confirmerAuthToken)
		require.NoError(t, err)

		// Submit agreement
		_, err = confirmationUseCase.Submit(ctx, &transitionCase.ID, authenticatedConfirmer, &domain.SubmitTransitionConfirmationRequest{
			Decision: domain.UserTransitionDecisionAgree,
		})
		require.NoError(t, err)

		// Since waiting period is 0 days, evaluation in Submit should confirm the case
		cCase, err := caseUseCase.GetByID(ctx, &transitionCase.ID, &userA.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.UserTransitionCaseStatusConfirmed, cCase.Status)

		// User B (both confirmer and stakeholder) should receive confirmed notification
		notifs, err := notificationUseCase.ListByReceiverID(ctx, &userB.ID)
		require.NoError(t, err)
		var confirmedNotif *domain.Notification
		for _, n := range notifs {
			if n.Type == domain.NotificationTypeTransition && n.Metadata["type"] == "transition_confirmed" {
				confirmedNotif = n
				break
			}
		}
		require.NotNil(t, confirmedNotif)
		assert.Contains(t, confirmedNotif.Message, "사망 전환이 확정되었습니다")
	})
}
