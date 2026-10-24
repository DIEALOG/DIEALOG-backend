package domain

import (
	"context"
	"time"

	"github.com/DIEALOG/DIEALOG-crypto/domain"
	"github.com/janghanul090801/collections"
)

type Legacy struct {
	ID        ID        `json:"id"`
	Name      string    `form:"name" binding:"required" json:"name"`
	UserID    ID        `json:"user_id"`
	Accounts  []Account `json:"accounts,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type LegacyResponse struct {
	ID        ID                    `json:"id"`
	Name      string                `json:"name"`
	UserID    ID                    `json:"user_id"`
	Accounts  []AccountListResponse `json:"accounts,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
}

func ToLegacyResponse(legacy Legacy) LegacyResponse {
	accounts := make([]AccountListResponse, 0, len(legacy.Accounts))

	for _, account := range legacy.Accounts {
		accounts = append(accounts, AccountListResponse{
			ID:       account.ID,
			LegacyID: account.LegacyID,
			Type:     account.Type,
			Name:     account.Name,
			Action:   account.Action,
			StakeholderIDs: collections.ToIterable(account.Stakeholders).Map(func(s Stakeholder) domain.ID {
				return s.ID
			}),
			StakeholderGroupIDs: collections.ToIterable(account.StakeholderGroups).Map(func(s StakeholderGroup) domain.ID {
				return s.ID
			}),
			CreatedAt: account.CreatedAt,
		})
	}
	return LegacyResponse{
		ID:        legacy.ID,
		Name:      legacy.Name,
		UserID:    legacy.UserID,
		Accounts:  accounts,
		CreatedAt: legacy.CreatedAt,
	}
}

type CreateAccountRequest struct {
	Type                AccountType   `json:"type" validate:"required"`
	Name                string        `json:"name" validate:"required"`
	AccountID           *string       `json:"account_id,omitempty"`
	AccountPassword     *string       `json:"account_password,omitempty"`
	DataLocation        *string       `json:"data_location,omitempty"`
	StakeholderIDs      []ID          `json:"stakeholder_ids,omitempty"`
	StakeholderGroupIDs []ID          `json:"stakeholder_group_ids,omitempty"`
	Action              AccountAction `json:"action" validate:"required"`
}

type UpdateAccountRequest struct {
	Type                AccountType   `json:"type" validate:"required"`
	Name                string        `json:"name" validate:"required"`
	AccountID           *string       `json:"account_id,omitempty"`
	AccountPassword     *string       `json:"account_password,omitempty"`
	DataLocation        *string       `json:"data_location,omitempty"`
	StakeholderIDs      []ID          `json:"stakeholder_ids,omitempty"`
	StakeholderGroupIDs []ID          `json:"stakeholder_group_ids,omitempty"`
	Action              AccountAction `json:"action" validate:"required"`
}

type Account struct {
	ID                ID                 `json:"id"`
	LegacyID          ID                 `json:"legacy_id"`
	Type              AccountType        `json:"type"`
	Name              string             `json:"name"`
	AccountID         *string            `json:"account_id,omitempty"`
	DataLocation      *string            `json:"data_location,omitempty"`
	Stakeholders      []Stakeholder      `json:"stakeholders,omitempty"`
	StakeholderGroups []StakeholderGroup `json:"stakeholder_groups,omitempty"`
	Action            AccountAction      `json:"action"`
	CreatedAt         time.Time          `json:"created_at"`
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
	case AccountActionDelete,
		AccountActionRemain,
		AccountActionPermissionMove,
		AccountActionBackupAndDelete,
		AccountActionOther:
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

type AccountListResponse struct {
	ID                  ID            `json:"id"`
	LegacyID            ID            `json:"legacy_id"`
	Type                AccountType   `json:"type"`
	Name                string        `json:"name"`
	Action              AccountAction `json:"action"`
	StakeholderIDs      []ID          `json:"stakeholder_ids,omitempty"`
	StakeholderGroupIDs []ID          `json:"stakeholder_group_ids,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
}

type AccountDetailResponse struct {
	ID                ID                 `json:"id"`
	LegacyID          ID                 `json:"legacy_id"`
	Type              AccountType        `json:"type"`
	Name              string             `json:"name"`
	AccountID         *string            `json:"account_id,omitempty"`
	AccountPassword   *string            `json:"account_password,omitempty"`
	DataLocation      *string            `json:"data_location,omitempty"`
	Stakeholders      []Stakeholder      `json:"stakeholders,omitempty"`
	StakeholderGroups []StakeholderGroup `json:"stakeholder_group_ids,omitempty"`
	Action            AccountAction      `json:"action"`
	CreatedAt         time.Time          `json:"created_at"`
}

func ToAccountDetailResponse(account Account, password *string) AccountDetailResponse {
	return AccountDetailResponse{
		ID:                account.ID,
		LegacyID:          account.LegacyID,
		Type:              account.Type,
		Name:              account.Name,
		AccountID:         account.AccountID,
		AccountPassword:   password,
		DataLocation:      account.DataLocation,
		Stakeholders:      account.Stakeholders,
		StakeholderGroups: account.StakeholderGroups,
		Action:            account.Action,
		CreatedAt:         account.CreatedAt,
	}
}

type LegacyRepository interface {
	Create(c context.Context, legacy *Legacy, userID *ID) (*Legacy, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*Legacy, error)
	FindByID(c context.Context, id *ID, userID *ID) (*Legacy, error)
	Update(c context.Context, id *ID, name string, userID *ID) (*Legacy, error)
	Delete(c context.Context, id *ID, userID *ID) error
	CreateAccount(c context.Context, account *Account, legacyID *ID, userID *ID) (*Account, error)
	FindAccountById(c context.Context, id *ID, userID *ID) (*Account, error)
	DeleteAccount(c context.Context, accountID *ID, legacyID *ID, userID *ID) error
}

type LegacyUseCase interface {
	Create(c context.Context, legacy *Legacy, userID *ID) (*Legacy, error)
	ListByUserID(c context.Context, userID *ID) ([]*Legacy, error)
	GetByID(c context.Context, id *ID, userID *ID) (*Legacy, error)
	Update(c context.Context, id *ID, name string, userID *ID) (*Legacy, error)
	Delete(c context.Context, id *ID, userID *ID) error
	CreateAccount(c context.Context, account *Account, password *string, legacyID *ID, userID *ID) (*Account, error)
	GetAccountByID(c context.Context, id *ID, userID *ID) (*Account, *string, error)
	DeleteAccount(c context.Context, accountID *ID, legacyID *ID, userID *ID) error
}
