package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/account"
	"github.com/DIEALOG/DIEALOG-backend/ent/legacy"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/janghanul090801/collections"
)

type legacyRepository struct {
	client *ent.Client
}

func NewLegacyRepository(client *ent.Client) domain.LegacyRepository {
	return &legacyRepository{
		client: client,
	}
}

func (r *legacyRepository) Create(c context.Context, leg *domain.Legacy, userID *domain.ID) (*domain.Legacy, error) {
	entity, err := r.client.Legacy.Create().
		SetName(leg.Name).
		SetOwnerID(*userID).
		Save(c)
	if err != nil {
		return nil, err
	}

	result := toDomainLegacy(entity)
	result.UserID = *userID
	return result, nil
}

func (r *legacyRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.Legacy, error) {
	query := r.client.Legacy.Query().
		Where(legacy.HasOwnerWith(user.IDEQ(*userID)))

	entities, err := query.
		WithAccounts(func(q *ent.AccountQuery) {
			q.WithStakeholders()
		}).
		Order(ent.Desc(legacy.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	legacies := make([]*domain.Legacy, len(entities))
	for i, entity := range entities {
		item := toDomainLegacy(entity)
		item.UserID = *userID
		legacies[i] = item
	}

	return legacies, nil
}

func (r *legacyRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.Legacy, error) {
	entity, err := r.client.Legacy.Query().
		Where(
			legacy.IDEQ(*id),
			legacy.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithAccounts(func(q *ent.AccountQuery) {
			q.WithStakeholders()
		}).
		Only(c)

	if err != nil {
		return nil, err
	}

	item := toDomainLegacy(entity)
	item.UserID = *userID
	return item, nil
}

func (r *legacyRepository) Update(c context.Context, id *domain.ID, name string, userID *domain.ID) (*domain.Legacy, error) {
	exists, err := r.client.Legacy.Query().
		Where(
			legacy.IDEQ(*id),
			legacy.HasOwnerWith(user.IDEQ(*userID)),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	entity, err := r.client.Legacy.UpdateOneID(*id).
		SetName(name).
		Save(c)
	if err != nil {
		return nil, err
	}

	item := toDomainLegacy(entity)
	item.UserID = *userID
	return item, nil
}

func (r *legacyRepository) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	n, err := r.client.Legacy.Delete().
		Where(
			legacy.IDEQ(*id),
			legacy.HasOwnerWith(user.IDEQ(*userID)),
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

func (r *legacyRepository) CreateAccount(c context.Context, acc *domain.Account, legacyID *domain.ID, userID *domain.ID) (*domain.Account, error) {
	stakeholderIDs := collections.ToIterable(acc.Stakeholders).Map(func(s domain.Stakeholder) domain.ID {
		return s.ID
	})

	stakeholderGroupIDs := collections.ToIterable(acc.StakeholderGroups).Map(func(s domain.StakeholderGroup) domain.ID {
		return s.ID
	})

	exists, err := r.client.Legacy.Query().
		Where(
			legacy.IDEQ(*legacyID),
			legacy.HasOwnerWith(user.IDEQ(*userID)),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ent.NotFoundError{}
	}

	builder := r.client.Account.Create().
		SetName(acc.Name).
		SetType(toEntAccountType(acc.Type.Type)).
		SetAction(toEntAccountAction(acc.Action.Action)).
		SetLegacyID(*legacyID)

	if acc.Type.Name != nil {
		builder.SetTypeName(*acc.Type.Name)
	}
	if acc.Action.Message != nil {
		builder.SetActionMessage(*acc.Action.Message)
	}
	if acc.DataLocation != nil {
		builder.SetDataLocation(*acc.DataLocation)
	}
	if acc.AccountID != nil {
		builder.SetAccountID(*acc.AccountID)
	}
	if len(stakeholderIDs) > 0 {
		builder.AddStakeholderIDs(stakeholderIDs...)
	}
	if len(stakeholderGroupIDs) > 0 {
		builder.AddStakeholderGroupIDs(stakeholderGroupIDs...)
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	loaded, err := r.client.Account.Query().
		Where(account.IDEQ(entity.ID)).
		WithStakeholders().
		Only(c)
	if err != nil {
		result := toDomainAccount(loaded)
		result.LegacyID = *legacyID
		return result, nil
	}

	result := toDomainAccount(loaded)
	result.LegacyID = *legacyID
	return result, nil
}

func (r *legacyRepository) FindAccountById(c context.Context, id *domain.ID, userID *domain.ID) (*domain.Account, error) {
	entity, err := r.client.Account.Query().
		Where(
			account.IDEQ(*id),
			account.HasLegacyWith(
				legacy.HasOwnerWith(user.IDEQ(*userID)),
			),
		).
		WithStakeholders().
		WithStakeholderGroups(func(q *ent.StakeholderGroupQuery) {
			q.WithStakeholders()
		}).
		WithLegacy().
		Only(c)

	if err != nil {
		return nil, err
	}

	return toDomainAccount(entity), nil
}

func (r *legacyRepository) DeleteAccount(c context.Context, accountID *domain.ID, legacyID *domain.ID, userID *domain.ID) error {
	n, err := r.client.Account.Delete().
		Where(
			account.IDEQ(*accountID),
			account.HasLegacyWith(
				legacy.HasOwnerWith(user.IDEQ(*userID)),
			),
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