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

	if legacy.Name == "" {
		return nil, domain.NewBadRequestError(errors.New("name is required"))
	}

	l, err := lu.legacyRepository.Create(ctx, legacy, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return l, nil
}

func (lu *legacyUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.Legacy, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	legacies, err := lu.legacyRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return legacies, nil
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

func (lu *legacyUseCase) CreateAccount(c context.Context, account *domain.Account, password *string, legacyID *domain.ID, userID *domain.ID) (*domain.Account, error) {
	if !account.IsValid() {
		return nil, domain.NewBadRequestError(errors.New("invalid account: type or action is invalid"))
	}

	if account.Name == "" {
		return nil, domain.NewBadRequestError(errors.New("invalid account: name is required"))
	}

	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	acc, err := lu.legacyRepository.CreateAccount(ctx, account, legacyID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	// Store the password only after the account row exists, so a failed
	// account creation cannot leave an orphaned credential behind.
	if password != nil {
		err := lu.credentialUseCase.StorePassword(ctx, userID, account.Name, *password)
		if err != nil {
			return nil, domain.NewInternalServerError(err)
		}
	}

	return acc, nil
}

func (lu *legacyUseCase) GetAccountPassword(c context.Context, userID *domain.ID, account *domain.Account) (*string, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	pwd, err := lu.credentialUseCase.GetPassword(ctx, userID, account.Name)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return pwd, nil
}

func (lu *legacyUseCase) GetAccountByID(c context.Context, legacyID *domain.ID, id *domain.ID, userID *domain.ID) (*domain.Account, *string, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	acc, err := lu.legacyRepository.FindAccountById(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, domain.NewNotFoundError(err)
		}
		return nil, nil, domain.NewInternalServerError(err)
	}

	// The route nests the account under a legacy; reject a mismatched parent.
	if acc.LegacyID != *legacyID {
		return nil, nil, domain.NewNotFoundError(errors.New("account not found"))
	}

	pwd, err := lu.credentialUseCase.GetPassword(ctx, userID, acc.Name)
	if err != nil {
		return nil, nil, domain.NewInternalServerError(err)
	}

	return acc, pwd, nil
}

func (lu *legacyUseCase) DeleteAccount(c context.Context, accountID *domain.ID, legacyID *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	// 계정의 패스워드 Credential 삭제 (삭제 전 패스워드 제거)
	acc, err := lu.legacyRepository.FindAccountById(ctx, accountID, userID)
	if err != nil {
		return domain.NewInternalServerError(err)
	}

	err = lu.credentialUseCase.DeletePassword(ctx, userID, acc.Name)
	if err != nil {
		return domain.NewInternalServerError(err)
	}

	// 계정 DB 삭제
	err = lu.legacyRepository.DeleteAccount(ctx, accountID, legacyID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}
