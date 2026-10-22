package repository_test

import (
	"context"
	"testing"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent/enttest"
	"github.com/DIEALOG/DIEALOG-backend/infra/repository"
	"github.com/stretchr/testify/assert"

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
		Name:      "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		Testament: "My Testament",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "My Testament", user.Testament)

	// 2. Create Stakeholder Group
	grp, err := groupRepo.Create(ctx, &domain.StakeholderGroup{
		Name: "Family",
	}, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, grp)
	assert.Equal(t, "Family", grp.Name)

	// 3. Create Stakeholder
	phone := "010-1234-5678"
	email := "sh@example.com"
	msg := "Thank you"
	sh, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
		Name:         "Stakeholder 1",
		Phone:        &phone,
		Email:        &email,
		Relationship: "Sister",
		Message:      &msg,
		GroupIDs:     []domain.ID{grp.ID},
	}, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, sh)
	assert.Equal(t, "Stakeholder 1", sh.Name)
	assert.Equal(t, "Sister", sh.Relationship)
	assert.Len(t, sh.Groups, 1)

	stakeholders, err := stakeholderRepo.FindAllByUserID(ctx, &user.ID)
	assert.NoError(t, err)
	assert.Len(t, stakeholders, 1)

	// 4. Create Legacy
	leg, err := legacyRepo.Create(ctx, &domain.Legacy{
		Name: "Social Accounts",
	}, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, leg)
	assert.Equal(t, "Social Accounts", leg.Name)

	// 5. Create Account under Legacy
	accID := "mygoogle@gmail.com"
	accPw := "secretpassword"
	actionMsg := "Please delete"
	acc, err := legacyRepo.CreateAccount(ctx, &domain.Account{
		Type: domain.AccountType{
			Type: domain.AccountTypeGoogle,
		},
		Action: domain.AccountAction{
			Action:  domain.AccountActionDelete,
			Message: &actionMsg,
		},
		AccountID:       &accID,
		AccountPassword: &accPw,
		StakeholderIDs:  []domain.ID{sh.ID},
	}, &leg.ID, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, domain.AccountTypeGoogle, acc.Type.Type)
	assert.Equal(t, domain.AccountActionDelete, acc.Action.Action)
	assert.Len(t, acc.Stakeholders, 1)

	// 6. Find Legacy by ID (with accounts)
	fetchedLeg, err := legacyRepo.FindByID(ctx, &leg.ID, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedLeg)
	assert.Len(t, fetchedLeg.Accounts, 1)
	assert.Equal(t, "mygoogle@gmail.com", *fetchedLeg.Accounts[0].AccountID)
	assert.Len(t, fetchedLeg.Accounts[0].Stakeholders, 1)

	// 7. Update Legacy
	updatedLeg, err := legacyRepo.Update(ctx, &leg.ID, "SNS Accounts", &user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "SNS Accounts", updatedLeg.Name)

	// 8. Delete Account
	err = legacyRepo.DeleteAccount(ctx, &acc.ID, &leg.ID, &user.ID)
	assert.NoError(t, err)

	// 9. Delete Legacy
	err = legacyRepo.Delete(ctx, &leg.ID, &user.ID)
	assert.NoError(t, err)

	// 10. Delete Stakeholder
	err = stakeholderRepo.Delete(ctx, &sh.ID, &user.ID)
	assert.NoError(t, err)

	// 11. Delete Group
	err = groupRepo.Delete(ctx, &grp.ID, &user.ID)
	assert.NoError(t, err)
}
