package domain

import (
	"context"
	"time"
)

type User struct {
	ID                ID                 `json:"id"`
	Name              string             `json:"name"`
	Email             string             `json:"email"`
	Password          string             `json:"-"`
	Stakeholders      []Stakeholder      `json:"stakeholders,omitempty"`
	StakeholderGroups []StakeholderGroup `json:"stakeholder_groups,omitempty"`
	Testament         string             `json:"testament"`
	CreatedAt         time.Time          `json:"created_at"`
}

type Stakeholder struct {
	ID           ID                 `json:"id"`
	UserID       ID                 `json:"user_id"`
	Name         string             `form:"name" binding:"required" json:"name"`
	Phone        *string            `form:"phone" json:"phone,omitempty"`
	Email        *string            `form:"email" json:"email,omitempty"`
	Relationship string             `form:"relationship" json:"relationship,omitempty"`
	Message      *string            `form:"message" json:"message,omitempty"`
	GroupIDs     []ID               `form:"group_ids" json:"group_ids,omitempty"`
	Groups       []StakeholderGroup `json:"groups,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

func (s Stakeholder) IsValid() bool {
	return s.Phone != nil || s.Email != nil
}

type StakeholderGroup struct {
	ID             ID            `json:"id"`
	UserID         ID            `json:"user_id"`
	Name           string        `form:"name" binding:"required" json:"name"`
	StakeholderIDs []ID          `form:"stakeholder_ids" json:"stakeholder_ids,omitempty"`
	Stakeholders   []Stakeholder `json:"stakeholders,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type UserRepository interface {
	Create(c context.Context, user *User) (*User, error)
	FindAll(c context.Context) ([]*User, error)
	FindByEmail(c context.Context, email string) (*User, error)
	FindByID(c context.Context, id *ID) (*User, error)
	UpdateTestament(c context.Context, id *ID, testament string) error
}

type StakeholderRepository interface {
	Create(c context.Context, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*Stakeholder, error)
	FindByID(c context.Context, id *ID, userID *ID) (*Stakeholder, error)
	Update(c context.Context, id *ID, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	Delete(c context.Context, id *ID, userID *ID) error
}

type StakeholderUseCase interface {
	Create(c context.Context, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	ListByUserID(c context.Context, userID *ID) ([]*Stakeholder, error)
	GetByID(c context.Context, id *ID, userID *ID) (*Stakeholder, error)
	Update(c context.Context, id *ID, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	Delete(c context.Context, id *ID, userID *ID) error
}

type StakeholderGroupRepository interface {
	Create(c context.Context, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*StakeholderGroup, error)
	FindByID(c context.Context, id *ID, userID *ID) (*StakeholderGroup, error)
	Update(c context.Context, id *ID, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	Delete(c context.Context, id *ID, userID *ID) error
}

type StakeholderGroupUseCase interface {
	Create(c context.Context, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	ListByUserID(c context.Context, userID *ID) ([]*StakeholderGroup, error)
	GetByID(c context.Context, id *ID, userID *ID) (*StakeholderGroup, error)
	Update(c context.Context, id *ID, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	Delete(c context.Context, id *ID, userID *ID) error
}
