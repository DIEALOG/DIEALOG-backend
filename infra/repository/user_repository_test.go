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

func TestCreate(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	repo := repository.NewUserRepository(client)

	user, err := repo.Create(context.Background(), &domain.User{
		Name:     "hanul",
		Email:    "hanul@gmail.com",
		Password: "123456",
	})

	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, "hanul", user.Name)
}
