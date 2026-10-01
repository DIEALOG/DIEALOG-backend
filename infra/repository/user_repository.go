package repository

import (
	"context"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/user"
)

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) domain.UserRepository {
	return &userRepository{
		client: client,
	}
}

func (r *userRepository) Create(c context.Context, user *domain.User) (*domain.User, error) {
	u, err := r.client.User.Create().
		SetName(user.Name).
		SetEmail(user.Email).
		SetPassword(user.Password).
		Save(c)
	if err != nil {
		return nil, err
	}

	return toDomainUser(u), nil
}

func (r *userRepository) FindAll(c context.Context) ([]*domain.User, error) {
	u, err := r.client.User.Query().All(c)
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(u))
	for i, v := range u {
		users[i] = toDomainUser(v)
	}

	return users, err
}

func (r *userRepository) FindByEmail(c context.Context, email string) (*domain.User, error) {
	u, err := r.client.User.Query().Where(
		user.EmailEQ(email),
	).Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainUser(u), err
}

func (r *userRepository) FindByID(c context.Context, id *domain.ID) (*domain.User, error) {
	u, err := r.client.User.Get(c, *id)
	if err != nil {
		return nil, err
	}

	return toDomainUser(u), err
}
