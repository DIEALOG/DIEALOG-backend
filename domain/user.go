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
	PhoneNumber       string             `json:"phone_number"`
	ProfileImagePath  *string            `json:"profile_image_path"`
	Stakeholders      []Stakeholder      `json:"stakeholders,omitempty"`
	StakeholderGroups []StakeholderGroup `json:"stakeholder_groups,omitempty"`
	Config            UserConfig         `json:"config"`
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
	if s.Name == "" {
		return false
	}
	phoneProvided := s.Phone != nil && *s.Phone != ""
	emailProvided := s.Email != nil && *s.Email != ""
	return phoneProvided || emailProvided
}

type StakeholderGroup struct {
	ID             ID            `json:"id"`
	UserID         ID            `json:"user_id"`
	Name           string        `form:"name" binding:"required" json:"name"`
	StakeholderIDs []ID          `form:"stakeholder_ids" json:"stakeholder_ids,omitempty"`
	Stakeholders   []Stakeholder `json:"stakeholders,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type UserConfig struct {
	ID                                      ID   `json:"id"`
	UserID                                  ID   `json:"user_id"`
	LockByFaceID                            bool `json:"lock_by_face_id"`
	CertificationBeforeSensitiveInformation bool `json:"certification_before_sensitive_information"`
}

type UserTransitionConfirmer struct {
	ID            ID  `json:"id"`
	UserID        ID  `json:"user_id"`
	StakeholderID *ID `json:"stakeholder_id,omitempty"`

	Name  string  `json:"name"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`

	Status     UserTransitionConfirmerStatus `json:"status"`
	InvitedAt  *time.Time                    `json:"invited_at,omitempty"`
	AcceptedAt *time.Time                    `json:"accepted_at,omitempty"`
	RevokedAt  *time.Time                    `json:"revoked_at,omitempty"`

	InviteTokenHash    string     `json:"-"`
	InviteExpiresAt    *time.Time `json:"invite_expires_at,omitempty"`
	AuthTokenHash      string     `json:"-"`
	AuthTokenExpiresAt *time.Time `json:"auth_token_expires_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserTransitionConfirmerStatus string

const (
	UserTransitionConfirmerStatusPending  UserTransitionConfirmerStatus = "pending"
	UserTransitionConfirmerStatusAccepted UserTransitionConfirmerStatus = "accepted"
	UserTransitionConfirmerStatusDeclined UserTransitionConfirmerStatus = "declined"
	UserTransitionConfirmerStatusRevoked  UserTransitionConfirmerStatus = "revoked"
	UserTransitionConfirmerStatusExpired  UserTransitionConfirmerStatus = "expired"
)

type UserTransitionPolicy struct {
	ID                     ID        `json:"id"`
	UserID                 ID        `json:"user_id"`
	MinimumConfirmers      int       `json:"minimum_confirmers"`
	WaitingPeriodDays      int       `json:"waiting_period_days"`
	AliveCheckIntervalDays int       `json:"alive_check_interval_days"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type UserAliveCheck struct {
	ID          ID                   `json:"id"`
	UserID      ID                   `json:"user_id"`
	Sequence    int                  `json:"sequence"`
	ScheduledAt time.Time            `json:"scheduled_at"`
	DueAt       time.Time            `json:"due_at"`
	RespondedAt *time.Time           `json:"responded_at,omitempty"`
	Status      UserAliveCheckStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
}

type UserAliveCheckStatus string

const (
	UserAliveCheckStatusScheduled UserAliveCheckStatus = "scheduled"
	UserAliveCheckStatusPending   UserAliveCheckStatus = "pending"
	UserAliveCheckStatusResponded UserAliveCheckStatus = "responded"
	UserAliveCheckStatusOverdue   UserAliveCheckStatus = "overdue"
	UserAliveCheckStatusCancelled UserAliveCheckStatus = "cancelled"
)

type UserTransitionCase struct {
	ID     ID `json:"id"`
	UserID ID `json:"user_id"`

	Status UserTransitionCaseStatus `json:"status"`

	MinimumConfirmersSnapshot int `json:"minimum_confirmers_snapshot"`
	WaitingPeriodDaysSnapshot int `json:"waiting_period_days_snapshot"`

	RequestedAt      time.Time  `json:"requested_at"`
	WaitingStartedAt *time.Time `json:"waiting_started_at,omitempty"`
	EligibleAt       *time.Time `json:"eligible_at,omitempty"`
	ConfirmedAt      *time.Time `json:"confirmed_at,omitempty"`
	CancelledAt      *time.Time `json:"cancelled_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserTransitionCaseStatus string

const (
	UserTransitionCaseStatusPending   UserTransitionCaseStatus = "pending_confirmation"
	UserTransitionCaseStatusWaiting   UserTransitionCaseStatus = "waiting_period"
	UserTransitionCaseStatusConfirmed UserTransitionCaseStatus = "confirmed"
	UserTransitionCaseStatusCancelled UserTransitionCaseStatus = "cancelled"
)

type UserTransitionConfirmation struct {
	ID                    ID `json:"id"`
	TransitionCaseID      ID `json:"transition_case_id"`
	TransitionConfirmerID ID `json:"transition_confirmer_id"`

	Decision    UserTransitionDecision `json:"decision"`
	Comment     *string                `json:"comment,omitempty"`
	RespondedAt *time.Time             `json:"responded_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type UserTransitionDecision string

const (
	UserTransitionDecisionAgree    UserTransitionDecision = "agree"
	UserTransitionDecisionDisagree UserTransitionDecision = "disagree"
	UserTransitionDecisionAbstain  UserTransitionDecision = "abstain"
)

type UserRepository interface {
	Create(c context.Context, user *User) (*User, error)
	FindAll(c context.Context) ([]*User, error)
	FindByEmail(c context.Context, email string) (*User, error)
	FindByID(c context.Context, id *ID) (*User, error)
	Update(c context.Context, user *User) (*User, error)
}

type StakeholderRepository interface {
	Create(c context.Context, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*Stakeholder, error)
	FindByID(c context.Context, id *ID, userID *ID) (*Stakeholder, error)
	Update(c context.Context, id *ID, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	Delete(c context.Context, id *ID, userID *ID) error
	FindAllByIDs(c context.Context, ids []ID, userID *ID) ([]*Stakeholder, error)
}

type StakeholderUseCase interface {
	Create(c context.Context, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	ListByUserID(c context.Context, userID *ID) ([]*Stakeholder, error)
	GetByID(c context.Context, id *ID, userID *ID) (*Stakeholder, error)
	Update(c context.Context, id *ID, stakeholder *Stakeholder, userID *ID) (*Stakeholder, error)
	Delete(c context.Context, id *ID, userID *ID) error
	ListByIDs(c context.Context, ids []ID, userID *ID) ([]*Stakeholder, error)
}

type StakeholderGroupRepository interface {
	Create(c context.Context, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*StakeholderGroup, error)
	FindByID(c context.Context, id *ID, userID *ID) (*StakeholderGroup, error)
	Update(c context.Context, id *ID, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	Delete(c context.Context, id *ID, userID *ID) error
	FindAllByIDs(c context.Context, ids []ID, userID *ID) ([]*StakeholderGroup, error)
}

type StakeholderGroupUseCase interface {
	Create(c context.Context, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	ListByUserID(c context.Context, userID *ID) ([]*StakeholderGroup, error)
	GetByID(c context.Context, id *ID, userID *ID) (*StakeholderGroup, error)
	Update(c context.Context, id *ID, group *StakeholderGroup, userID *ID) (*StakeholderGroup, error)
	Delete(c context.Context, id *ID, userID *ID) error
	ListByIDs(c context.Context, ids []ID, userID *ID) ([]*StakeholderGroup, error)
}

type UserConfigRepository interface {
	Create(c context.Context, config *UserConfig) (*UserConfig, error)
	FindByUserID(c context.Context, userID *ID) (*UserConfig, error)
	Update(c context.Context, config *UserConfig) (*UserConfig, error)
}
