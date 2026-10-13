package repository

import (
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent/account"
)

func toDomainUser(entity *ent.User) *domain.User {
	u := &domain.User{
		ID:        entity.ID,
		Name:      entity.Name,
		Email:     entity.Email,
		Password:  entity.Password,
		Testament: entity.Testament,
		CreatedAt: entity.CreatedAt,
	}

	if entity.Edges.Stakeholders != nil {
		u.Stakeholders = make([]domain.Stakeholder, len(entity.Edges.Stakeholders))
		for i, s := range entity.Edges.Stakeholders {
			u.Stakeholders[i] = *toDomainStakeholder(s)
		}
	}

	return u
}

func toDomainStakeholder(entity *ent.Stakeholder) *domain.Stakeholder {
	s := &domain.Stakeholder{
		ID:        entity.ID,
		Name:      entity.Name,
		Phone:     entity.Phone,
		Email:     entity.Email,
		CreatedAt: entity.CreatedAt,
	}
	if entity.Edges.Owner != nil {
		s.UserID = entity.Edges.Owner.ID
	}
	return s
}

func toDomainLegacy(entity *ent.Legacy) *domain.Legacy {
	l := &domain.Legacy{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
	}
	if entity.Edges.Owner != nil {
		l.UserID = entity.Edges.Owner.ID
	}
	if entity.Edges.Accounts != nil {
		l.Accounts = make([]domain.Account, len(entity.Edges.Accounts))
		for i, a := range entity.Edges.Accounts {
			l.Accounts[i] = *toDomainAccount(a)
		}
	}
	return l
}

func toDomainAccount(entity *ent.Account) *domain.Account {
	a := &domain.Account{
		ID:               entity.ID,
		Type:             domain.AccountType(entity.Type),
		AccountTypeOther: entity.AccountTypeOther,
		AccountID:        entity.AccountID,
		AccountPassword:  entity.AccountPassword,
		CreatedAt:        entity.CreatedAt,
	}
	if entity.Edges.Legacy != nil {
		a.LegacyID = entity.Edges.Legacy.ID
	}
	return a
}

func toEntAccountType(t domain.AccountType) account.Type {
	return account.Type(t)
}
