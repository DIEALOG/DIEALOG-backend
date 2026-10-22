package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/stakeholder"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
)

type stakeholderRepository struct {
	client *ent.Client
}

func NewStakeholderRepository(client *ent.Client) domain.StakeholderRepository {
	return &stakeholderRepository{
		client: client,
	}
}

func (r *stakeholderRepository) Create(c context.Context, s *domain.Stakeholder, userID *domain.ID) (*domain.Stakeholder, error) {
	builder := r.client.Stakeholder.Create().
		SetName(s.Name).
		SetOwnerID(*userID).
		SetRelationship(s.Relationship)

	if s.Phone != nil {
		builder.SetPhone(*s.Phone)
	}
	if s.Email != nil {
		builder.SetEmail(*s.Email)
	}
	if s.Message != nil {
		builder.SetMessage(*s.Message)
	}
	if len(s.GroupIDs) > 0 {
		builder.AddGroupIDs(s.GroupIDs...)
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	// reload with groups
	loaded, err := r.client.Stakeholder.Query().
		Where(stakeholder.IDEQ(entity.ID)).
		WithGroups().
		Only(c)
	if err != nil {
		result := toDomainStakeholder(entity)
		result.UserID = *userID
		return result, nil
	}

	result := toDomainStakeholder(loaded)
	result.UserID = *userID
	return result, nil
}

func (r *stakeholderRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.Stakeholder, error) {
	entities, err := r.client.Stakeholder.Query().
		Where(stakeholder.HasOwnerWith(user.IDEQ(*userID))).
		WithGroups().
		Order(ent.Asc(stakeholder.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.Stakeholder, len(entities))
	for i, entity := range entities {
		item := toDomainStakeholder(entity)
		item.UserID = *userID
		results[i] = item
	}

	return results, nil
}

func (r *stakeholderRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.Stakeholder, error) {
	entity, err := r.client.Stakeholder.Query().
		Where(
			stakeholder.IDEQ(*id),
			stakeholder.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithGroups().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainStakeholder(entity)
	item.UserID = *userID
	return item, nil
}

func (r *stakeholderRepository) Update(c context.Context, id *domain.ID, s *domain.Stakeholder, userID *domain.ID) (*domain.Stakeholder, error) {
	exists, err := r.client.Stakeholder.Query().
		Where(
			stakeholder.IDEQ(*id),
			stakeholder.HasOwnerWith(user.IDEQ(*userID)),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	builder := r.client.Stakeholder.UpdateOneID(*id).
		SetName(s.Name).
		SetRelationship(s.Relationship)

	if s.Phone != nil {
		builder.SetPhone(*s.Phone)
	} else {
		builder.ClearPhone()
	}

	if s.Email != nil {
		builder.SetEmail(*s.Email)
	} else {
		builder.ClearEmail()
	}

	if s.Message != nil {
		builder.SetMessage(*s.Message)
	} else {
		builder.ClearMessage()
	}

	if s.GroupIDs != nil {
		builder.ClearGroups()
		if len(s.GroupIDs) > 0 {
			builder.AddGroupIDs(s.GroupIDs...)
		}
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	loaded, err := r.client.Stakeholder.Query().
		Where(stakeholder.IDEQ(entity.ID)).
		WithGroups().
		Only(c)
	if err != nil {
		result := toDomainStakeholder(entity)
		result.UserID = *userID
		return result, nil
	}

	result := toDomainStakeholder(loaded)
	result.UserID = *userID
	return result, nil
}

func (r *stakeholderRepository) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	n, err := r.client.Stakeholder.Delete().
		Where(
			stakeholder.IDEQ(*id),
			stakeholder.HasOwnerWith(user.IDEQ(*userID)),
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
