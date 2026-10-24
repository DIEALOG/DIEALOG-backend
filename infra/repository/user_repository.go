package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
)

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) domain.UserRepository {
	return &userRepository{
		client: client,
	}
}

func (r *userRepository) Create(c context.Context, u *domain.User) (*domain.User, error) {
	create := r.client.User.Create().
		SetName(u.Name).
		SetEmail(u.Email).
		SetPassword(u.Password).
		SetPhoneNumber(u.PhoneNumber)

	saved, err := create.Save(c)
	if err != nil {
		return nil, err
	}

	return toDomainUser(saved), nil
}

func (r *userRepository) FindAll(c context.Context) ([]*domain.User, error) {
	u, err := r.client.User.Query().WithStakeholders().All(c)
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
	).WithStakeholders().Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainUser(u), err
}

func (r *userRepository) FindByID(c context.Context, id *domain.ID) (*domain.User, error) {
	u, err := r.client.User.Query().Where(
		user.IDEQ(*id),
	).WithStakeholders().Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainUser(u), err
}

func (r *userRepository) Update(c context.Context, u *domain.User) (*domain.User, error) {
	exist, err := r.client.User.Query().Where(
		user.IDEQ(u.ID),
	).Exist(c)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, &ent.NotFoundError{}
	}

	updated, err := r.client.User.UpdateOneID(u.ID).
		SetName(u.Name).
		SetEmail(u.Email).
		SetPassword(u.Password).
		SetPhoneNumber(u.PhoneNumber).
		SetNillableProfileImagePath(u.ProfileImagePath).
		Save(c)
	if err != nil {
		return nil, err
	}

	return toDomainUser(updated), nil
}
