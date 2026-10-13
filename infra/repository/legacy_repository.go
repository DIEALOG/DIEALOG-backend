package repository

import (
	"context"

	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/account"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/legacy"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/user"
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

func (r *legacyRepository) FindAllByUserID(c context.Context, userID *domain.ID, offset, limit int) ([]*domain.Legacy, int, error) {
	query := r.client.Legacy.Query().
		Where(legacy.HasOwnerWith(user.IDEQ(*userID)))

	total, err := query.Count(c)
	if err != nil {
		return nil, 0, err
	}

	entities, err := query.
		WithAccounts().
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(legacy.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, 0, err
	}

	legacies := make([]*domain.Legacy, len(entities))
	for i, entity := range entities {
		item := toDomainLegacy(entity)
		item.UserID = *userID
		legacies[i] = item
	}

	return legacies, total, nil
}

func (r *legacyRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.Legacy, error) {
	entity, err := r.client.Legacy.Query().
		Where(
			legacy.IDEQ(*id),
			legacy.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithAccounts().
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
		SetType(toEntAccountType(acc.Type)).
		SetAccountID(acc.AccountID).
		SetAccountPassword(acc.AccountPassword).
		SetLegacyID(*legacyID)

	if acc.AccountTypeOther != nil {
		builder.SetAccountTypeOther(*acc.AccountTypeOther)
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	result := toDomainAccount(entity)
	result.LegacyID = *legacyID
	return result, nil
}

func (r *legacyRepository) DeleteAccount(c context.Context, accountID *domain.ID, legacyID *domain.ID, userID *domain.ID) error {
	n, err := r.client.Account.Delete().
		Where(
			account.IDEQ(*accountID),
			account.HasLegacyWith(
				legacy.IDEQ(*legacyID),
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
