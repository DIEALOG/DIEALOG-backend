package repository_test

import (
	"context"
	"testing"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/enttest"
	"github.com/DIEALOG/DIEALOG-backend/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestLegacyRepository(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	legacyRepo := repository.NewLegacyRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)
	groupRepo := repository.NewStakeholderGroupRepository(client)

	ctx := context.Background()

	// 1. Create User
	user, err := userRepo.Create(ctx, &domain.User{
		Name:        "testuser",
		Email:       "test@example.com",
		Password:    "password123",
		PhoneNumber: "010-1234-5678",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	// 2. Create Stakeholder Group
	grp, err := groupRepo.Create(ctx, &domain.StakeholderGroup{
		Name: "Family",
	}, &user.ID)
	require.NoError(t, err)
	require.NotNil(t, grp)
	assert.Equal(t, "Family", grp.Name)

	// 3. Create Stakeholder
	phone := "010-1234-5678"
	email := "sh@example.com"
	sh, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
		Name:         "Stakeholder 1",
		Phone:        &phone,
		Email:        &email,
		Relationship: "Sister",
		GroupIDs:     []domain.ID{grp.ID},
	}, &user.ID)
	require.NoError(t, err)
	require.NotNil(t, sh)
	assert.Equal(t, "Stakeholder 1", sh.Name)
	assert.Equal(t, "Sister", sh.Relationship)
	assert.Len(t, sh.Groups, 1)

	stakeholders, err := stakeholderRepo.FindAllByUserID(ctx, &user.ID)
	require.NoError(t, err)
	assert.Len(t, stakeholders, 1)

	// 4. Create Legacy
	leg, err := legacyRepo.Create(ctx, &domain.Legacy{
		Name: "Social Accounts",
	}, &user.ID)
	require.NoError(t, err)
	require.NotNil(t, leg)
	assert.Equal(t, "Social Accounts", leg.Name)

	// 5. Create Account under Legacy
	accID := "mygoogle@gmail.com"
	actionMsg := "Please delete"
	acc, err := legacyRepo.CreateAccount(ctx, &domain.Account{
		Name: "Google",
		Type: domain.AccountType{
			Type: domain.AccountTypeGoogle,
		},
		Action: domain.AccountAction{
			Action:  domain.AccountActionDelete,
			Message: &actionMsg,
		},
		AccountID:    &accID,
		Stakeholders: []domain.Stakeholder{*sh},
	}, &leg.ID, &user.ID)
	require.NoError(t, err)
	require.NotNil(t, acc)
	assert.Equal(t, domain.AccountTypeGoogle, acc.Type.Type)
	assert.Equal(t, domain.AccountActionDelete, acc.Action.Action)
	assert.Len(t, acc.Stakeholders, 1)

	// 6. Find Legacy by ID (with accounts)
	fetchedLeg, err := legacyRepo.FindByID(ctx, &leg.ID, &user.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedLeg)
	require.Len(t, fetchedLeg.Accounts, 1)
	assert.Equal(t, "mygoogle@gmail.com", *fetchedLeg.Accounts[0].AccountID)
	assert.Len(t, fetchedLeg.Accounts[0].Stakeholders, 1)

	// 7. Update Legacy
	updatedLeg, err := legacyRepo.Update(ctx, &leg.ID, "SNS Accounts", &user.ID)
	require.NoError(t, err)
	assert.Equal(t, "SNS Accounts", updatedLeg.Name)

	// 8. Delete Account
	err = legacyRepo.DeleteAccount(ctx, &acc.ID, &leg.ID, &user.ID)
	require.NoError(t, err)

	// 9. Delete Legacy
	err = legacyRepo.Delete(ctx, &leg.ID, &user.ID)
	require.NoError(t, err)

	// 10. Delete Stakeholder
	err = stakeholderRepo.Delete(ctx, &sh.ID, &user.ID)
	require.NoError(t, err)

	// 11. Delete Group
	err = groupRepo.Delete(ctx, &grp.ID, &user.ID)
	require.NoError(t, err)
}

func TestLegacyRepositoryUpdateAccount(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:update_account?mode=memory&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	legacyRepo := repository.NewLegacyRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)
	groupRepo := repository.NewStakeholderGroupRepository(client)

	ctx := context.Background()

	owner, err := userRepo.Create(ctx, &domain.User{
		Name:        "owner",
		Email:       "update-owner@example.com",
		Password:    "password",
		PhoneNumber: "010-5555-1111",
	})
	require.NoError(t, err)
	other, err := userRepo.Create(ctx, &domain.User{
		Name:        "other",
		Email:       "update-other@example.com",
		Password:    "password",
		PhoneNumber: "010-5555-2222",
	})
	require.NoError(t, err)

	firstEmail := "first@example.com"
	first, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{Name: "First", Email: &firstEmail}, &owner.ID)
	require.NoError(t, err)
	secondEmail := "second@example.com"
	second, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{Name: "Second", Email: &secondEmail}, &owner.ID)
	require.NoError(t, err)
	group, err := groupRepo.Create(ctx, &domain.StakeholderGroup{
		Name:           "Handlers",
		StakeholderIDs: []domain.ID{first.ID},
	}, &owner.ID)
	require.NoError(t, err)

	leg, err := legacyRepo.Create(ctx, &domain.Legacy{Name: "Social"}, &owner.ID)
	require.NoError(t, err)

	actionMessage := "delete after death"
	accountID := "account-1"
	dataLocation := "drive"
	acc, err := legacyRepo.CreateAccount(ctx, &domain.Account{
		Name:              "Google",
		Type:              domain.AccountType{Type: domain.AccountTypeGoogle},
		Action:            domain.AccountAction{Action: domain.AccountActionDelete, Message: &actionMessage},
		AccountID:         &accountID,
		DataLocation:      &dataLocation,
		Stakeholders:      []domain.Stakeholder{*first},
		StakeholderGroups: []domain.StakeholderGroup{*group},
	}, &leg.ID, &owner.ID)
	require.NoError(t, err)
	require.Len(t, acc.Stakeholders, 1)
	require.Len(t, acc.StakeholderGroups, 1)

	// The update replaces stakeholders/groups instead of only adding to them,
	// and clears optional fields that are no longer provided.
	updated, err := legacyRepo.UpdateAccount(ctx, &domain.Account{
		ID:           acc.ID,
		Name:         "KakaoTalk",
		Type:         domain.AccountType{Type: domain.AccountTypeKakaoTalk},
		Action:       domain.AccountAction{Action: domain.AccountActionRemain},
		Stakeholders: []domain.Stakeholder{*second},
	}, &leg.ID, &owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "KakaoTalk", updated.Name)
	assert.Equal(t, domain.AccountTypeKakaoTalk, updated.Type.Type)
	assert.Equal(t, domain.AccountActionRemain, updated.Action.Action)
	assert.Nil(t, updated.AccountID)
	assert.Nil(t, updated.DataLocation)
	assert.Nil(t, updated.Action.Message)
	assert.Equal(t, leg.ID, updated.LegacyID)
	require.Len(t, updated.Stakeholders, 1)
	assert.Equal(t, second.ID, updated.Stakeholders[0].ID)
	assert.Empty(t, updated.StakeholderGroups)

	// Another user cannot update the account.
	_, err = legacyRepo.UpdateAccount(ctx, &domain.Account{
		ID:     acc.ID,
		Name:   "hijacked",
		Type:   domain.AccountType{Type: domain.AccountTypeGoogle},
		Action: domain.AccountAction{Action: domain.AccountActionDelete},
	}, &leg.ID, &other.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	// The owner cannot update the account through a legacy they do not own.
	otherLegacy, err := legacyRepo.Create(ctx, &domain.Legacy{Name: "Other"}, &other.ID)
	require.NoError(t, err)
	_, err = legacyRepo.UpdateAccount(ctx, &domain.Account{
		ID:     acc.ID,
		Name:   "moved",
		Type:   domain.AccountType{Type: domain.AccountTypeGoogle},
		Action: domain.AccountAction{Action: domain.AccountActionDelete},
	}, &otherLegacy.ID, &owner.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
