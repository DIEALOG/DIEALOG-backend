package domain

import (
	"context"
	"time"
)

type Legacy struct {
	ID        ID        `json:"id"`
	Name      string    `form:"name" binding:"required" json:"name"`
	UserID    ID        `json:"user_id"`
	Accounts  []Account `json:"accounts,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID               ID          `json:"id"`
	LegacyID         ID          `json:"legacy_id"`
	Type             AccountType `form:"type" binding:"required" json:"type"`
	AccountTypeOther *string     `form:"account_type_other" json:"account_type_other,omitempty"`
	AccountID        string      `form:"account_id" binding:"required" json:"account_id"`
	AccountPassword  string      `form:"account_password" binding:"required" json:"account_password"`
	CreatedAt        time.Time   `json:"created_at"`
}

func (a Account) IsValid() bool {
	return (a.Type != AccountTypeOther) || (a.AccountTypeOther != nil && a.Type == AccountTypeOther)
}

type AccountType string

const (
	AccountTypeGoogle    AccountType = "Google"
	AccountTypeFacebook  AccountType = "Facebook"
	AccountTypeTwitter   AccountType = "Twitter"
	AccountTypeApple     AccountType = "Apple"
	AccountTypeDiscord   AccountType = "Discord"
	AccountTypeMicrosoft AccountType = "Microsoft"
	AccountTypeKakaoTalk AccountType = "KakaoTalk"
	AccountTypeGitHub    AccountType = "GitHub"
	AccountTypeOther     AccountType = "Other"
)

type LegacyRepository interface {
	Create(c context.Context, legacy *Legacy, userID *ID) (*Legacy, error)
	FindAllByUserID(c context.Context, userID *ID, offset, limit int) ([]*Legacy, int, error)
	FindByID(c context.Context, id *ID, userID *ID) (*Legacy, error)
	Update(c context.Context, id *ID, name string, userID *ID) (*Legacy, error)
	Delete(c context.Context, id *ID, userID *ID) error
	CreateAccount(c context.Context, account *Account, legacyID *ID, userID *ID) (*Account, error)
	DeleteAccount(c context.Context, accountID *ID, legacyID *ID, userID *ID) error
}

type LegacyUseCase interface {
	Create(c context.Context, legacy *Legacy, userID *ID) (*Legacy, error)
	ListByUserID(c context.Context, userID *ID, offset, limit int) (*PaginatedResult[*Legacy], error)
	GetByID(c context.Context, id *ID, userID *ID) (*Legacy, error)
	Update(c context.Context, id *ID, name string, userID *ID) (*Legacy, error)
	Delete(c context.Context, id *ID, userID *ID) error
	CreateAccount(c context.Context, account *Account, legacyID *ID, userID *ID) (*Account, error)
	DeleteAccount(c context.Context, accountID *ID, legacyID *ID, userID *ID) error
}
