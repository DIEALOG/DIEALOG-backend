package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/userconfig"
)

type userConfigRepository struct {
	client *ent.Client
}

func NewUserConfigRepository(client *ent.Client) domain.UserConfigRepository {
	return &userConfigRepository{
		client: client,
	}
}

func (r *userConfigRepository) Create(
	c context.Context,
	config *domain.UserConfig,
) (*domain.UserConfig, error) {
	create := r.client.UserConfig.Create().
		SetUserID(config.UserID)

	saved, err := create.Save(c)
	if err != nil {
		return nil, err
	}

	return toDomainUserConfig(saved), nil
}

func (r *userConfigRepository) FindByUserID(
	c context.Context,
	userID *domain.ID,
) (*domain.UserConfig, error) {
	config, err := r.client.UserConfig.Query().
		Where(userconfig.HasUserWith(user.IDEQ(*userID))).
		Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainUserConfig(config), nil
}

func (r *userConfigRepository) Update(
	c context.Context,
	config *domain.UserConfig,
) (*domain.UserConfig, error) {
	exists, err := r.client.UserConfig.Query().
		Where(userconfig.HasUserWith(user.IDEQ(config.UserID))).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	_, saveErr := r.client.UserConfig.Update().
		Where(userconfig.HasUserWith(user.IDEQ(config.UserID))).
		SetLockByFaceID(config.LockByFaceID).
		SetCertificationBeforeSensitiveInformation(
			config.CertificationBeforeSensitiveInformation,
		).
		Save(c)
	if saveErr != nil {
		return nil, saveErr
	}

	result2, err2 := r.client.UserConfig.Query().
		Where(userconfig.HasUserWith(user.IDEQ(config.UserID))).
		Only(c)
	if err2 != nil {
		return nil, err2
	}

	return toDomainUserConfig(result2), nil
}