package repository_test

import (
	"context"
	"testing"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent/enttest"
	"github.com/DIEALOG/DIEALOG-backend/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// TestStakeholderRepositoryLinksExistingUser verifies that registering a
// stakeholder whose email or phone belongs to an existing user stores that
// user's ID in stakeholder_id.
func TestStakeholderRepositoryLinksExistingUser(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:stakeholder_link?mode=memory&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)

	ctx := context.Background()

	owner, err := userRepo.Create(ctx, &domain.User{
		Name:        "owner",
		Email:       "owner@example.com",
		Password:    "password123",
		PhoneNumber: "010-0000-0000",
	})
	require.NoError(t, err)

	member, err := userRepo.Create(ctx, &domain.User{
		Name:        "member",
		Email:       "member@example.com",
		Password:    "password123",
		PhoneNumber: "010-2222-2222",
	})
	require.NoError(t, err)

	t.Run("matches by email", func(t *testing.T) {
		email := "member@example.com"
		created, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "by email",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)
		require.NotNil(t, created.StakeholderID)
		assert.Equal(t, member.ID, *created.StakeholderID)
	})

	t.Run("matches email ignoring case and spaces", func(t *testing.T) {
		email := "  MEMBER@Example.com "
		created, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "by email variant",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)
		require.NotNil(t, created.StakeholderID)
		assert.Equal(t, member.ID, *created.StakeholderID)
	})

	t.Run("matches by phone", func(t *testing.T) {
		phone := "010-2222-2222"
		created, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "by phone",
			Phone: &phone,
		}, &owner.ID)
		require.NoError(t, err)
		require.NotNil(t, created.StakeholderID)
		assert.Equal(t, member.ID, *created.StakeholderID)
	})

	t.Run("no matching user", func(t *testing.T) {
		email := "stranger@example.com"
		phone := "010-9999-9999"
		created, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "stranger",
			Email: &email,
			Phone: &phone,
		}, &owner.ID)
		require.NoError(t, err)
		assert.Nil(t, created.StakeholderID)
	})

	t.Run("update relinks and clears", func(t *testing.T) {
		email := "stranger@example.com"
		created, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
			Name:  "relink",
			Email: &email,
		}, &owner.ID)
		require.NoError(t, err)
		require.Nil(t, created.StakeholderID)

		matched := "member@example.com"
		updated, err := stakeholderRepo.Update(ctx, &created.ID, &domain.Stakeholder{
			Name:  "relink",
			Email: &matched,
		}, &owner.ID)
		require.NoError(t, err)
		require.NotNil(t, updated.StakeholderID)
		assert.Equal(t, member.ID, *updated.StakeholderID)

		unmatched := "stranger@example.com"
		cleared, err := stakeholderRepo.Update(ctx, &created.ID, &domain.Stakeholder{
			Name:  "relink",
			Email: &unmatched,
		}, &owner.ID)
		require.NoError(t, err)
		assert.Nil(t, cleared.StakeholderID)
	})
}
