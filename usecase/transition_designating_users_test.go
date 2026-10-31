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

func TestListDesignatingUsers(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:designating_users?mode=memory&_fk=1")
	defer client.Close()

	ctx := context.Background()
	userRepo := repository.NewUserRepository(client)
	confirmerRepo := repository.NewUserTransitionConfirmerRepository(client)
	confirmerUseCase := usecase.NewUserTransitionConfirmerUseCase(
		confirmerRepo,
		nil,
		userRepo,
		nil,
		2*time.Second,
	)

	me, err := userRepo.Create(ctx, &domain.User{
		Password:    "password",
		Name:        "me",
		Email:       "me@example.com",
		PhoneNumber: "010-1111-2222",
	})
	require.NoError(t, err)

	emailOwner, err := userRepo.Create(ctx, &domain.User{
		Password:    "password",
		Name:        "email owner",
		Email:       "email-owner@example.com",
		PhoneNumber: "010-3333-4444",
	})
	require.NoError(t, err)

	phoneOwner, err := userRepo.Create(ctx, &domain.User{
		Password:    "password",
		Name:        "phone owner",
		Email:       "phone-owner@example.com",
		PhoneNumber: "010-5555-6666",
	})
	require.NoError(t, err)

	stranger, err := userRepo.Create(ctx, &domain.User{
		Password:    "password",
		Name:        "stranger",
		Email:       "stranger@example.com",
		PhoneNumber: "010-7777-8888",
	})
	require.NoError(t, err)

	revokedOwner, err := userRepo.Create(ctx, &domain.User{
		Password:    "password",
		Name:        "revoked owner",
		Email:       "revoked-owner@example.com",
		PhoneNumber: "010-9999-0000",
	})
	require.NoError(t, err)

	// The same person is designated by email and by phone from different owners.
	myEmail := me.Email
	_, err = confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
		Name:  "me",
		Email: &myEmail,
	}, &emailOwner.ID)
	require.NoError(t, err)

	myPhone := me.PhoneNumber
	_, err = confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
		Name:  "me",
		Phone: &myPhone,
	}, &phoneOwner.ID)
	require.NoError(t, err)

	// Unrelated designation must not be returned.
	strangerEmail := stranger.Email
	_, err = confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
		Name:  "stranger",
		Email: &strangerEmail,
	}, &stranger.ID)
	require.NoError(t, err)

	// A revoked designation is no longer active and must be excluded.
	revoked, err := confirmerUseCase.Invite(ctx, &domain.InviteTransitionConfirmerRequest{
		Name:  "me",
		Email: &myEmail,
	}, &revokedOwner.ID)
	require.NoError(t, err)
	require.NoError(t, confirmerUseCase.Revoke(ctx, &revoked.Confirmer.ID, &revokedOwner.ID))

	items, err := confirmerUseCase.ListDesignatingUsers(ctx, &me.ID)
	require.NoError(t, err)
	require.Len(t, items, 2)

	ids := make([]domain.ID, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	assert.ElementsMatch(t, []domain.ID{emailOwner.ID, phoneOwner.ID}, ids)
	assert.NotContains(t, ids, me.ID)
}
