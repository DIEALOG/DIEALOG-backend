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

	if entity.Edges.StakeholderGroups != nil {
		u.StakeholderGroups = make([]domain.StakeholderGroup, len(entity.Edges.StakeholderGroups))
		for i, g := range entity.Edges.StakeholderGroups {
			u.StakeholderGroups[i] = *toDomainStakeholderGroup(g)
		}
	}

	return u
}

func toDomainStakeholder(entity *ent.Stakeholder) *domain.Stakeholder {
	s := &domain.Stakeholder{
		ID:           entity.ID,
		Name:         entity.Name,
		Phone:        entity.Phone,
		Email:        entity.Email,
		Relationship: entity.Relationship,
		Message:      entity.Message,
		CreatedAt:    entity.CreatedAt,
	}
	if entity.Edges.Owner != nil {
		s.UserID = entity.Edges.Owner.ID
	}
	if entity.Edges.Groups != nil {
		s.Groups = make([]domain.StakeholderGroup, len(entity.Edges.Groups))
		s.GroupIDs = make([]domain.ID, len(entity.Edges.Groups))
		for i, g := range entity.Edges.Groups {
			s.Groups[i] = *toDomainStakeholderGroup(g)
			s.GroupIDs[i] = g.ID
		}
	}
	return s
}

func toDomainStakeholderGroup(entity *ent.StakeholderGroup) *domain.StakeholderGroup {
	g := &domain.StakeholderGroup{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
	}
	if entity.Edges.Owner != nil {
		g.UserID = entity.Edges.Owner.ID
	}
	if entity.Edges.Stakeholders != nil {
		g.Stakeholders = make([]domain.Stakeholder, len(entity.Edges.Stakeholders))
		g.StakeholderIDs = make([]domain.ID, len(entity.Edges.Stakeholders))
		for i, s := range entity.Edges.Stakeholders {
			g.Stakeholders[i] = *toDomainStakeholder(s)
			g.StakeholderIDs[i] = s.ID
		}
	}
	return g
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
		ID: entity.ID,
		Type: domain.AccountType{
			ID:   entity.ID,
			Type: domain.AccountTypeField(entity.Type),
			Name: entity.TypeName,
		},
		AccountID:       entity.AccountID,
		AccountPassword: entity.AccountPassword,
		DataLocation:    entity.DataLocation,
		Action: domain.AccountAction{
			ID:      entity.ID,
			Action:  domain.AccountActionField(entity.Action),
			Message: entity.ActionMessage,
		},
		CreatedAt: entity.CreatedAt,
	}
	if entity.Edges.Legacy != nil {
		a.LegacyID = entity.Edges.Legacy.ID
	}
	if entity.Edges.Stakeholders != nil {
		a.Stakeholders = make([]domain.Stakeholder, len(entity.Edges.Stakeholders))
		a.StakeholderIDs = make([]domain.ID, len(entity.Edges.Stakeholders))
		for i, s := range entity.Edges.Stakeholders {
			a.Stakeholders[i] = *toDomainStakeholder(s)
			a.StakeholderIDs[i] = s.ID
		}
	}
	return a
}

func toEntAccountType(t domain.AccountTypeField) account.Type {
	return account.Type(t)
}

func toEntAccountAction(a domain.AccountActionField) account.Action {
	if a == "" {
		return account.ActionRemain
	}
	return account.Action(a)
}
