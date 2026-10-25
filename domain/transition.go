package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrActiveTransitionCaseExists = errors.New("an active transition case already exists")
	ErrTransitionCaseNotEligible  = errors.New("transition case is not eligible for confirmation")
	ErrInvalidTransitionState     = errors.New("invalid transition state")
	ErrInvitationExpired          = errors.New("invitation expired")
	ErrConfirmerNotAccepted       = errors.New("confirmer invitation has not been accepted")
)

type UpdateUserTransitionPolicyRequest struct {
	MinimumConfirmers      *int `json:"minimum_confirmers"`
	WaitingPeriodDays      *int `json:"waiting_period_days"`
	AliveCheckIntervalDays *int `json:"alive_check_interval_days"`
}

type InviteTransitionConfirmerRequest struct {
	StakeholderID *ID     `json:"stakeholder_id"`
	Name          string  `json:"name"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
}

type TransitionInvitationResponse struct {
	Confirmer       *UserTransitionConfirmer `json:"confirmer"`
	InvitationToken string                   `json:"invitation_token"`
	ExpiresAt       time.Time                `json:"expires_at"`
}

type AcceptTransitionInvitationResponse struct {
	Confirmer *UserTransitionConfirmer `json:"confirmer"`
	AuthToken string                   `json:"auth_token"`
	ExpiresAt time.Time                `json:"expires_at"`
}

type SubmitTransitionConfirmationRequest struct {
	Decision UserTransitionDecision `json:"decision"`
	Comment  *string                `json:"comment,omitempty"`
}

type TransitionInvitationTokenRequest struct {
	Token string `json:"token"`
}

type UserTransitionPolicyRepository interface {
	FindByUserID(c context.Context, userID *ID) (*UserTransitionPolicy, error)
	Create(c context.Context, policy *UserTransitionPolicy, userID *ID) (*UserTransitionPolicy, error)
	Update(c context.Context, policy *UserTransitionPolicy, userID *ID) (*UserTransitionPolicy, error)
}

type UserTransitionPolicyUseCase interface {
	Get(c context.Context, userID *ID) (*UserTransitionPolicy, error)
	Update(c context.Context, request *UpdateUserTransitionPolicyRequest, userID *ID) (*UserTransitionPolicy, error)
}

type UserTransitionConfirmerRepository interface {
	Create(c context.Context, confirmer *UserTransitionConfirmer, userID *ID) (*UserTransitionConfirmer, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*UserTransitionConfirmer, error)
	FindByID(c context.Context, id *ID, userID *ID) (*UserTransitionConfirmer, error)
	FindByInviteTokenHash(c context.Context, tokenHash string) (*UserTransitionConfirmer, error)
	FindByAuthTokenHash(c context.Context, tokenHash string) (*UserTransitionConfirmer, error)
	Update(c context.Context, confirmer *UserTransitionConfirmer, userID *ID) (*UserTransitionConfirmer, error)
}

type UserTransitionConfirmerUseCase interface {
	Invite(c context.Context, request *InviteTransitionConfirmerRequest, userID *ID) (*TransitionInvitationResponse, error)
	ListByUserID(c context.Context, userID *ID) ([]*UserTransitionConfirmer, error)
	Resend(c context.Context, id *ID, userID *ID) (*TransitionInvitationResponse, error)
	Revoke(c context.Context, id *ID, userID *ID) error
	Accept(c context.Context, token string) (*AcceptTransitionInvitationResponse, error)
	Decline(c context.Context, token string) error
	Authenticate(c context.Context, token string) (*UserTransitionConfirmer, error)
}

type UserAliveCheckRepository interface {
	Create(c context.Context, check *UserAliveCheck, userID *ID) (*UserAliveCheck, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*UserAliveCheck, error)
	FindByID(c context.Context, id *ID, userID *ID) (*UserAliveCheck, error)
	FindLatestByUserID(c context.Context, userID *ID) (*UserAliveCheck, error)
	Update(c context.Context, check *UserAliveCheck, userID *ID) (*UserAliveCheck, error)
}

type UserAliveCheckUseCase interface {
	ListByUserID(c context.Context, userID *ID) ([]*UserAliveCheck, error)
	ScheduleNext(c context.Context, userID *ID) (*UserAliveCheck, error)
	Respond(c context.Context, id *ID, userID *ID) (*UserAliveCheck, error)
}

type UserTransitionCaseRepository interface {
	Create(c context.Context, transitionCase *UserTransitionCase, userID *ID) (*UserTransitionCase, error)
	FindByID(c context.Context, id *ID, userID *ID) (*UserTransitionCase, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*UserTransitionCase, error)
	FindActiveByUserID(c context.Context, userID *ID) (*UserTransitionCase, error)
	Update(c context.Context, transitionCase *UserTransitionCase, userID *ID) (*UserTransitionCase, error)
	ConfirmAndCreateTasks(c context.Context, id *ID, userID *ID, now time.Time) (*UserTransitionCase, error)
}

type UserTransitionCaseUseCase interface {
	Start(c context.Context, confirmer *UserTransitionConfirmer) (*UserTransitionCase, error)
	GetByID(c context.Context, id *ID, userID *ID) (*UserTransitionCase, error)
	ListByUserID(c context.Context, userID *ID) ([]*UserTransitionCase, error)
	ListActiveByUserID(c context.Context, userID *ID) ([]*UserTransitionCase, error)
	Cancel(c context.Context, id *ID, userID *ID) (*UserTransitionCase, error)
	Evaluate(c context.Context, id *ID, userID *ID) (*UserTransitionCase, error)
}

type UserTransitionConfirmationRepository interface {
	Upsert(c context.Context, confirmation *UserTransitionConfirmation, userID *ID) (*UserTransitionConfirmation, error)
	FindByCaseID(c context.Context, caseID *ID, userID *ID) ([]*UserTransitionConfirmation, error)
	CountAgreeByCaseID(c context.Context, caseID *ID, userID *ID) (int, error)
}

type UserTransitionConfirmationUseCase interface {
	Submit(c context.Context, caseID *ID, confirmer *UserTransitionConfirmer, request *SubmitTransitionConfirmationRequest) (*UserTransitionConfirmation, error)
	ListByCaseID(c context.Context, caseID *ID, userID *ID) ([]*UserTransitionConfirmation, error)
}

type AccountHandlingTaskRepository interface {
	FindAllByCaseID(c context.Context, caseID *ID, userID *ID) ([]*AccountHandlingTask, error)
	FindByID(c context.Context, id *ID, userID *ID) (*AccountHandlingTask, error)
	FindByAssigneeTokenHash(c context.Context, tokenHash string) (*AccountHandlingTask, *AccountHandlingTaskAssignee, error)
	EnsureAssignees(c context.Context, taskID *ID, userID *ID, assignees []AccountHandlingTaskAssignee) (*AccountHandlingTask, error)
	SaveRequest(c context.Context, task *AccountHandlingTask, assignees []AccountHandlingTaskAssignee, userID *ID) (*AccountHandlingTask, error)
	SaveResult(c context.Context, task *AccountHandlingTask, assignee *AccountHandlingTaskAssignee, userID *ID) (*AccountHandlingTask, error)
}

type AccountHandlingEventRepository interface {
	Create(c context.Context, event *AccountHandlingEvent, userID *ID) (*AccountHandlingEvent, error)
	FindByTaskID(c context.Context, taskID *ID, userID *ID) ([]*AccountHandlingEvent, error)
}

type AccountHandlingUseCase interface {
	ListByCaseID(c context.Context, caseID *ID, userID *ID) ([]*AccountHandlingTask, error)
	Request(c context.Context, taskID *ID, userID *ID) (*AccountHandlingRequestResponse, error)
	SubmitResult(c context.Context, request *SubmitAccountHandlingResultRequest) (*AccountHandlingTask, error)
}
