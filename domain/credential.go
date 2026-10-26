package domain

import (
	"context"

	crypto "github.com/DIEALOG/DIEALOG-crypto/domain"
)

type StoreCredentialRequest struct {
	crypto.StoreCredentialRequest
}

type GetPasswordRequest struct {
	crypto.GetPasswordRequest
}

type DeleteCredentialRequest struct {
	crypto.DeletePasswordRequest
}

type CryptoServiceClient interface {
	Store(ctx context.Context, dto StoreCredentialRequest) error
	Retrieve(ctx context.Context, dto GetPasswordRequest) (*string, error)
	Delete(ctx context.Context, dto DeleteCredentialRequest) error
}

type CredentialUseCase interface {
	StorePassword(ctx context.Context, userID *ID, accountName, password string) error
	GetPassword(ctx context.Context, userID *ID, accountName string) (*string, error)
	DeletePassword(ctx context.Context, userID *ID, accountName string) error
}
