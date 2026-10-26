package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/stakeholder"
	stakeholdergrouppkg "github.com/DIEALOG/DIEALOG-backend/ent/stakeholdergroup"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/janghanul090801/collections"
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
		stakeholderIDs, err := r.ownedStakeholderIDs(c, g.StakeholderIDs, userID)
		if err != nil {
			return nil, err
		}
		if len(stakeholderIDs) > 0 {
			builder.AddStakeholderIDs(stakeholderIDs...)
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
		stakeholderIDs, err := r.ownedStakeholderIDs(c, g.StakeholderIDs, userID)
		if err != nil {
			return nil, err
		}
		builder.ClearStakeholders()
		if len(stakeholderIDs) > 0 {
			builder.AddStakeholderIDs(stakeholderIDs...)
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

func (r *stakeholderGroupRepository) FindAllByIDs(c context.Context, ids []domain.ID, userID *domain.ID) ([]*domain.StakeholderGroup, error) {
	stakeholders, err := r.client.StakeholderGroup.Query().
		Where(
			stakeholdergrouppkg.IDIn(ids...),
			stakeholdergrouppkg.HasOwnerWith(
				user.IDEQ(*userID),
			),
		).
		WithOwner().
		WithAccounts().
		WithStakeholders().
		All(c)
	if err != nil {
		return nil, err
	}

	return collections.ToIterable(stakeholders).Map(toDomainStakeholderGroup), nil
}

// ownedStakeholderIDs filters the client-supplied stakeholder IDs down to the
// ones owned by userID. Requesting a stakeholder that does not exist or belongs
// to another user is rejected so that a caller can never link someone else's
// stakeholder into their group.
func (r *stakeholderGroupRepository) ownedStakeholderIDs(c context.Context, ids []domain.ID, userID *domain.ID) ([]domain.ID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	owned, err := r.client.Stakeholder.Query().
		Where(
			stakeholder.IDIn(ids...),
			stakeholder.HasOwnerWith(user.IDEQ(*userID)),
		).
		IDs(c)
	if err != nil {
		return nil, err
	}

	distinct := make(map[domain.ID]struct{}, len(ids))
	for _, id := range ids {
		distinct[id] = struct{}{}
	}
	if len(owned) != len(distinct) {
		return nil, &ent.NotFoundError{}
	}

	return owned, nil
}
