package repository

import (
	"context"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
	stakeholdergrouppkg "github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/stakeholdergroup"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/user"
)

type stakeholderGroupRepository struct {
	client *ent.Client
}

func NewStakeholderGroupRepository(client *ent.Client) domain.StakeholderGroupRepository {
	return &stakeholderGroupRepository{
		client: client,
	}
}

func (r *stakeholderGroupRepository) Create(c context.Context, g *domain.StakeholderGroup, userID *domain.ID) (*domain.StakeholderGroup, error) {
	builder := r.client.StakeholderGroup.Create().
		SetName(g.Name).
		SetOwnerID(*userID)

	if len(g.StakeholderIDs) > 0 {
		builder.AddStakeholderIDs(g.StakeholderIDs...)
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	loaded, err := r.client.StakeholderGroup.Query().
		Where(stakeholdergrouppkg.IDEQ(entity.ID)).
		WithStakeholders().
		Only(c)
	if err != nil {
		result := toDomainStakeholderGroup(entity)
		result.UserID = *userID
		return result, nil
	}

	result := toDomainStakeholderGroup(loaded)
	result.UserID = *userID
	return result, nil
}

func (r *stakeholderGroupRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.StakeholderGroup, error) {
	entities, err := r.client.StakeholderGroup.Query().
		Where(stakeholdergrouppkg.HasOwnerWith(user.IDEQ(*userID))).
		WithStakeholders().
		Order(ent.Asc(stakeholdergrouppkg.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.StakeholderGroup, len(entities))
	for i, entity := range entities {
		item := toDomainStakeholderGroup(entity)
		item.UserID = *userID
		results[i] = item
	}

	return results, nil
}

func (r *stakeholderGroupRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.StakeholderGroup, error) {
	entity, err := r.client.StakeholderGroup.Query().
		Where(
			stakeholdergrouppkg.IDEQ(*id),
			stakeholdergrouppkg.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithStakeholders().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainStakeholderGroup(entity)
	item.UserID = *userID
	return item, nil
}

func (r *stakeholderGroupRepository) Update(c context.Context, id *domain.ID, g *domain.StakeholderGroup, userID *domain.ID) (*domain.StakeholderGroup, error) {
	exists, err := r.client.StakeholderGroup.Query().
		Where(
			stakeholdergrouppkg.IDEQ(*id),
			stakeholdergrouppkg.HasOwnerWith(user.IDEQ(*userID)),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	builder := r.client.StakeholderGroup.UpdateOneID(*id).
		SetName(g.Name)

	if g.StakeholderIDs != nil {
		builder.ClearStakeholders()
		if len(g.StakeholderIDs) > 0 {
			builder.AddStakeholderIDs(g.StakeholderIDs...)
		}
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	loaded, err := r.client.StakeholderGroup.Query().
		Where(stakeholdergrouppkg.IDEQ(entity.ID)).
		WithStakeholders().
		Only(c)
	if err != nil {
		result := toDomainStakeholderGroup(entity)
		result.UserID = *userID
		return result, nil
	}

	result := toDomainStakeholderGroup(loaded)
	result.UserID = *userID
	return result, nil
}

func (r *stakeholderGroupRepository) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	n, err := r.client.StakeholderGroup.Delete().
		Where(
			stakeholdergrouppkg.IDEQ(*id),
			stakeholdergrouppkg.HasOwnerWith(user.IDEQ(*userID)),
		).
		Exec(c)
	if err != nil {
		return err
	}
	if n == 0 {
		return &ent.NotFoundError{}
	}

	return nil
}
