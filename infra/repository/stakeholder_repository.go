package repository

import (
	"context"
	"strings"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/stakeholder"
	"github.com/DIEALOG/DIEALOG-backend/ent/stakeholdergroup"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/janghanul090801/collections"
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

	linkedUserID, err := r.linkedUserID(c, s.Email, s.Phone)
	if err != nil {
		return nil, err
	}
	builder.SetNillableStakeholderID(linkedUserID)

	if len(s.GroupIDs) > 0 {
		groupIDs, err := r.ownedGroupIDs(c, s.GroupIDs, userID)
		if err != nil {
			return nil, err
		}
		if len(groupIDs) > 0 {
			builder.AddGroupIDs(groupIDs...)
		}
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

	linkedUserID, err := r.linkedUserID(c, s.Email, s.Phone)
	if err != nil {
		return nil, err
	}
	if linkedUserID != nil {
		builder.SetStakeholderID(*linkedUserID)
	} else {
		builder.ClearStakeholderID()
	}

	if s.GroupIDs != nil {
		groupIDs, err := r.ownedGroupIDs(c, s.GroupIDs, userID)
		if err != nil {
			return nil, err
		}
		builder.ClearGroups()
		if len(groupIDs) > 0 {
			builder.AddGroupIDs(groupIDs...)
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

func (r *stakeholderRepository) FindAllByIDs(c context.Context, ids []domain.ID, userID *domain.ID) ([]*domain.Stakeholder, error) {
	stakeholders, err := r.client.Stakeholder.Query().
		Where(
			stakeholder.IDIn(ids...),
			stakeholder.HasOwnerWith(
				user.IDEQ(*userID),
			),
		).
		WithOwner().
		WithAccounts().
		WithGroups().
		All(c)
	if err != nil {
		return nil, err
	}

	return collections.ToIterable(stakeholders).Map(toDomainStakeholder), nil
}

// linkedUserID resolves the stakeholder's contact details to an existing
// DIEALOG user. A stakeholder is not required to be a member, so a lookup miss
// simply returns nil.
func (r *stakeholderRepository) linkedUserID(c context.Context, email *string, phone *string) (*domain.ID, error) {
	if email != nil {
		if normalized := strings.TrimSpace(*email); normalized != "" {
			found, err := r.client.User.Query().
				Where(user.EmailEqualFold(normalized)).
				Order(ent.Asc(user.FieldCreatedAt)).
				First(c)
			if err == nil {
				id := domain.ID(found.ID)
				return &id, nil
			}
			if !ent.IsNotFound(err) {
				return nil, err
			}
		}
	}

	if phone != nil {
		if normalized := strings.TrimSpace(*phone); normalized != "" {
			found, err := r.client.User.Query().
				Where(user.PhoneNumberEQ(normalized)).
				Order(ent.Asc(user.FieldCreatedAt)).
				First(c)
			if err == nil {
				id := domain.ID(found.ID)
				return &id, nil
			}
			if !ent.IsNotFound(err) {
				return nil, err
			}
		}
	}

	return nil, nil
}

// ownedGroupIDs filters the client-supplied group IDs down to the ones owned by
// userID. Requesting a group that does not exist or belongs to another user is
// rejected so that a caller can never link someone else's stakeholder group.
func (r *stakeholderRepository) ownedGroupIDs(c context.Context, ids []domain.ID, userID *domain.ID) ([]domain.ID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	owned, err := r.client.StakeholderGroup.Query().
		Where(
			stakeholdergroup.IDIn(ids...),
			stakeholdergroup.HasOwnerWith(user.IDEQ(*userID)),
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
