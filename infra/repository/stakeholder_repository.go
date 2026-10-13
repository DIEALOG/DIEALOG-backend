package repository

import (
	"context"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/stakeholder"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/user"
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
		SetOwnerID(*userID)

	if s.Phone != nil {
		builder.SetPhone(*s.Phone)
	}
	if s.Email != nil {
		builder.SetEmail(*s.Email)
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	result := toDomainStakeholder(entity)
	result.UserID = *userID
	return result, nil
}

func (r *stakeholderRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.Stakeholder, error) {
	entities, err := r.client.Stakeholder.Query().
		Where(stakeholder.HasOwnerWith(user.IDEQ(*userID))).
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
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainStakeholder(entity)
	item.UserID = *userID
	return item, nil
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
