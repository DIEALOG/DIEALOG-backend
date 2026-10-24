package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type legacyUseCase struct {
	legacyRepository  domain.LegacyRepository
	credentialUseCase domain.CredentialUseCase
	contextTimeout    time.Duration
}

func NewLegacyUseCase(legacyRepository domain.LegacyRepository, credentialUseCase domain.CredentialUseCase, timeout time.Duration) domain.LegacyUseCase {
	return &legacyUseCase{
		legacyRepository:  legacyRepository,
		credentialUseCase: credentialUseCase,
		contextTimeout:    timeout,
	}
}

func (lu *legacyUseCase) Create(c context.Context, legacy *domain.Legacy, userID *domain.ID) (*domain.Legacy, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	l, err := lu.legacyRepository.Create(ctx, legacy, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return l, nil
}

func (lu *legacyUseCase) ListByUserID(c context.Context, userID *domain.ID, offset, limit int) (*domain.PaginatedResult[*domain.Legacy], error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	legacies, total, err := lu.legacyRepository.FindAllByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.PaginatedResult[*domain.Legacy]{
		Items:  legacies,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (lu *legacyUseCase) GetByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.Legacy, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	l, err := lu.legacyRepository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return l, nil
}

func (lu *legacyUseCase) Update(c context.Context, id *domain.ID, name string, userID *domain.ID) (*domain.Legacy, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	l, err := lu.legacyRepository.Update(ctx, id, name, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return l, nil
}

func (lu *legacyUseCase) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	err := lu.legacyRepository.Delete(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}

func (lu *legacyUseCase) CreateAccount(c context.Context, account *domain.Account, password string, legacyID *domain.ID, userID *domain.ID) (*domain.Account, error) {
	if !account.IsValid() {
		return nil, domain.NewBadRequestError(errors.New("invalid account: type or action is invalid"))
	}

	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	err := lu.credentialUseCase.StorePassword(ctx, userID, account.Name, password)

	acc, err := lu.legacyRepository.CreateAccount(ctx, account, legacyID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return acc, nil
}

func (lu *legacyUseCase) GetAccountPassword(c context.Context, userID *domain.ID, account *domain.Account) (string, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	pwd, err := lu.credentialUseCase.GetPassword(ctx, userID, account.Name)
	if err != nil {
		return "", domain.NewInternalServerError(err)
	}

	return pwd, nil
}

func (lu *legacyUseCase) DeleteAccount(c context.Context, accountID *domain.ID, legacyID *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	err := lu.legacyRepository.DeleteAccount(ctx, accountID, legacyID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}
