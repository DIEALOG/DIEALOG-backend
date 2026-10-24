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

func (u *credentialUseCase) StorePassword(ctx context.Context, userID *domain.ID, accountName string, password string) error {
	return u.cryptoClient.Store(ctx, domain.StoreCredentialRequest{
		UserID:      *userID,
		AccountName: accountName,
		Password:    password,
	})
}

func (u *credentialUseCase) GetPassword(ctx context.Context, userID *domain.ID, accountName string) (*string, error) {
	return u.cryptoClient.Retrieve(ctx, domain.GetPasswordRequest{
		UserID:      *userID,
		AccountName: accountName,
	})
}
