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
	ID              ID            `json:"id"`
	LegacyID        ID            `json:"legacy_id"`
	Type            AccountType   `json:"type"`
	AccountID       *string       `form:"account_id" json:"account_id,omitempty"`
	AccountPassword *string       `form:"account_password" json:"account_password,omitempty"`
	DataLocation    *string       `form:"data_location" json:"data_location,omitempty"`
	StakeholderIDs  []ID          `form:"stakeholder_ids" json:"stakeholder_ids,omitempty"`
	Stakeholders    []Stakeholder `json:"stakeholders,omitempty"`
	Action          AccountAction `json:"action"`
	CreatedAt       time.Time     `json:"created_at"`
}

func (a Account) IsValid() bool {
	return a.Type.IsValid() && a.Action.IsValid()
}

type AccountType struct {
	ID   ID               `json:"id"`
	Type AccountTypeField `json:"type"`
	Name *string          `json:"name,omitempty"`
}

func (a AccountType) IsValid() bool {
	if a.Type == AccountTypeOther {
		return a.Name != nil && *a.Name != ""
	}
	switch a.Type {
	case AccountTypeGoogle, AccountTypeFacebook, AccountTypeTwitter, AccountTypeApple,
		AccountTypeDiscord, AccountTypeMicrosoft, AccountTypeKakaoTalk, AccountTypeGitHub,
		AccountTypeTelegram, AccountTypeInstagram:
		return true
	default:
		return false
	}
}

type AccountAction struct {
	ID      ID                 `json:"id"`
	Action  AccountActionField `json:"action"`
	Message *string            `json:"message,omitempty"`
}

func (a AccountAction) IsValid() bool {
	switch a.Action {
	case AccountActionDelete, AccountActionRemain, AccountActionPermissionMove,
		AccountActionBackupAndDelete, AccountActionOther:
		return true
	case "":
		return true
	default:
		return false
	}
}

type AccountTypeField string

const (
	AccountTypeGoogle    AccountTypeField = "Google"
	AccountTypeFacebook  AccountTypeField = "Facebook"
	AccountTypeTwitter   AccountTypeField = "Twitter"
	AccountTypeApple     AccountTypeField = "Apple"
	AccountTypeDiscord   AccountTypeField = "Discord"
	AccountTypeMicrosoft AccountTypeField = "Microsoft"
	AccountTypeKakaoTalk AccountTypeField = "KakaoTalk"
	AccountTypeGitHub    AccountTypeField = "GitHub"
	AccountTypeTelegram  AccountTypeField = "Telegram"
	AccountTypeInstagram AccountTypeField = "Instagram"
	AccountTypeOther     AccountTypeField = "Other"
)

type AccountActionField string

const (
	AccountActionDelete          AccountActionField = "delete"
	AccountActionRemain          AccountActionField = "remain"
	AccountActionPermissionMove  AccountActionField = "permission-move"
	AccountActionBackupAndDelete AccountActionField = "backup-and-delete"
	AccountActionOther           AccountActionField = "other"
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
