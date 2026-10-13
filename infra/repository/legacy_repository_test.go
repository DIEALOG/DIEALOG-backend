package repository_test

import (
	"context"
	"testing"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/enttest"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/infra/repository"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

func TestLegacyRepository(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	legacyRepo := repository.NewLegacyRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)

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

	// 2. Create Stakeholder
	phone := "010-1234-5678"
	email := "sh@example.com"
	sh, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
		Name:  "Stakeholder 1",
		Phone: &phone,
		Email: &email,
	}, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, sh)
	assert.Equal(t, "Stakeholder 1", sh.Name)

	stakeholders, err := stakeholderRepo.FindAllByUserID(ctx, &user.ID)
	assert.NoError(t, err)
	assert.Len(t, stakeholders, 1)

	// 3. Create Legacy
	leg, err := legacyRepo.Create(ctx, &domain.Legacy{
		Name: "Social Accounts",
	}, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, leg)
	assert.Equal(t, "Social Accounts", leg.Name)

	// 4. Create Account under Legacy
	acc, err := legacyRepo.CreateAccount(ctx, &domain.Account{
		Type:            domain.AccountTypeGoogle,
		AccountID:       "mygoogle@gmail.com",
		AccountPassword: "secretpassword",
	}, &leg.ID, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, domain.AccountTypeGoogle, acc.Type)

	// 5. Find Legacy by ID (with accounts)
	fetchedLeg, err := legacyRepo.FindByID(ctx, &leg.ID, &user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedLeg)
	assert.Len(t, fetchedLeg.Accounts, 1)
	assert.Equal(t, "mygoogle@gmail.com", fetchedLeg.Accounts[0].AccountID)

	// 6. Update Legacy
	updatedLeg, err := legacyRepo.Update(ctx, &leg.ID, "SNS Accounts", &user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "SNS Accounts", updatedLeg.Name)

	// 7. Delete Account
	err = legacyRepo.DeleteAccount(ctx, &acc.ID, &leg.ID, &user.ID)
	assert.NoError(t, err)

	// 8. Delete Legacy
	err = legacyRepo.Delete(ctx, &leg.ID, &user.ID)
	assert.NoError(t, err)

	// 9. Delete Stakeholder
	err = stakeholderRepo.Delete(ctx, &sh.ID, &user.ID)
	assert.NoError(t, err)
}
