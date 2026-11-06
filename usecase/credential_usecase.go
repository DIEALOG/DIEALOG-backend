package usecase

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
)

type credentialUseCase struct {
	cryptoClient domain.CryptoServiceClient
}

func NewCredentialUseCase(cryptoClient domain.CryptoServiceClient) domain.CredentialUseCase {
	return &credentialUseCase{cryptoClient: cryptoClient}
}

func (u *credentialUseCase) StorePassword(ctx context.Context, userID *domain.ID, accountID *domain.ID, password string) error {
	return u.cryptoClient.Store(ctx, domain.StoreCredentialRequest{
		UserID:    *userID,
		AccountID: *accountID,
		Password:  password,
	})
}

func (u *credentialUseCase) GetPassword(ctx context.Context, userID *domain.ID, accountID *domain.ID) (*string, error) {
	return u.cryptoClient.Retrieve(ctx, domain.GetPasswordRequest{
		UserID:    *userID,
		AccountID: *accountID,
	})
}

func (u *credentialUseCase) DeletePassword(ctx context.Context, userID *domain.ID, accountID *domain.ID) error {
	return u.cryptoClient.Delete(ctx, domain.DeleteCredentialRequest{
		UserID:    *userID,
		AccountID: *accountID,
	})
}
