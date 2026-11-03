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

func TestLastWordRepositoryFindAllByUserIDAndReceiverID(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:last_word_receiver?mode=memory&_fk=1")
	defer client.Close()

	userRepo := repository.NewUserRepository(client)
	stakeholderRepo := repository.NewStakeholderRepository(client)
	groupRepo := repository.NewStakeholderGroupRepository(client)
	lastWordRepo := repository.NewLastWordRepository(client)

	ctx := context.Background()

	owner, err := userRepo.Create(ctx, &domain.User{
		Name:        "owner",
		Email:       "owner@example.com",
		Password:    "password",
		PhoneNumber: "010-1111-1111",
	})
	require.NoError(t, err)

	receiver, err := userRepo.Create(ctx, &domain.User{
		Name:        "receiver",
		Email:       "receiver@example.com",
		Password:    "password",
		PhoneNumber: "010-2222-2222",
	})
	require.NoError(t, err)

	stranger, err := userRepo.Create(ctx, &domain.User{
		Name:        "stranger",
		Email:       "stranger@example.com",
		Password:    "password",
		PhoneNumber: "010-3333-3333",
	})
	require.NoError(t, err)

	// Stakeholder.Create links stakeholder_id from a matching email/phone.
	receiverEmail := receiver.Email
	receiverContact, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
		Name:  "Receiver",
		Email: &receiverEmail,
	}, &owner.ID)
	require.NoError(t, err)
	require.NotNil(t, receiverContact.StakeholderID)
	assert.Equal(t, receiver.ID, *receiverContact.StakeholderID)

	strangerEmail := stranger.Email
	strangerContact, err := stakeholderRepo.Create(ctx, &domain.Stakeholder{
		Name:  "Stranger",
		Email: &strangerEmail,
	}, &owner.ID)
	require.NoError(t, err)

	group, err := groupRepo.Create(ctx, &domain.StakeholderGroup{
		Name:           "Family",
		StakeholderIDs: []domain.ID{receiverContact.ID},
	}, &owner.ID)
	require.NoError(t, err)

	direct, err := lastWordRepo.Create(ctx, &domain.LastWord{
		UserID:   owner.ID,
		Title:    "direct",
		Message:  "hello",
		Receiver: []domain.Stakeholder{*receiverContact},
	})
	require.NoError(t, err)

	viaGroup, err := lastWordRepo.Create(ctx, &domain.LastWord{
		UserID:         owner.ID,
		Title:          "group",
		Message:        "hello",
		ReceiverGroups: []domain.StakeholderGroup{*group},
	})
	require.NoError(t, err)

	_, err = lastWordRepo.Create(ctx, &domain.LastWord{
		UserID:   owner.ID,
		Title:    "stranger",
		Message:  "hello",
		Receiver: []domain.Stakeholder{*strangerContact},
	})
	require.NoError(t, err)

	received, err := lastWordRepo.FindAllByUserIDAndReceiverID(ctx, &owner.ID, &receiver.ID)
	require.NoError(t, err)
	require.Len(t, received, 2)

	ids := make(map[domain.ID]bool, len(received))
	for _, lw := range received {
		ids[lw.ID] = true
	}
	assert.True(t, ids[direct.ID])
	assert.True(t, ids[viaGroup.ID])

	strangerReceived, err := lastWordRepo.FindAllByUserIDAndReceiverID(ctx, &owner.ID, &stranger.ID)
	require.NoError(t, err)
	require.Len(t, strangerReceived, 1)
	assert.Equal(t, "stranger", strangerReceived[0].Title)

	// A last word is only visible to receivers of the given author.
	_, err = lastWordRepo.FindAllByUserIDAndReceiverID(ctx, &receiver.ID, &receiver.ID)
	require.NoError(t, err)
}

func TestLastWordRepositoryFindAllByUserIDAndReceiverIDNotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:last_word_receiver_empty?mode=memory&_fk=1")
	defer client.Close()

	lastWordRepo := repository.NewLastWordRepository(client)

	lastWords, err := lastWordRepo.FindAllByUserIDAndReceiverID(context.Background(), &domain.ID{}, &domain.ID{})
	require.NoError(t, err)
	assert.Empty(t, lastWords)
}
