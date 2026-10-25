package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/internal/securetoken"
)

const (
	confirmerInvitationTTL = 7 * 24 * time.Hour
	confirmerAuthTTL       = 30 * 24 * time.Hour
)

type userTransitionConfirmerUseCase struct {
	repository            domain.UserTransitionConfirmerRepository
	stakeholderRepository domain.StakeholderRepository
	contextTimeout        time.Duration
}

func NewUserTransitionConfirmerUseCase(repository domain.UserTransitionConfirmerRepository, stakeholderRepository domain.StakeholderRepository, timeout time.Duration) domain.UserTransitionConfirmerUseCase {
	return &userTransitionConfirmerUseCase{
		repository:            repository,
		stakeholderRepository: stakeholderRepository,
		contextTimeout:        timeout,
	}
}

func (u *userTransitionConfirmerUseCase) Invite(c context.Context, request *domain.InviteTransitionConfirmerRequest, userID *domain.ID) (*domain.TransitionInvitationResponse, error) {
	if request == nil {
		return nil, domain.NewBadRequestError(errors.New("request is required"))
	}

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	name := strings.TrimSpace(request.Name)
	email := request.Email
	phone := request.Phone

	if request.StakeholderID != nil {
		stakeholder, err := u.stakeholderRepository.FindByID(ctx, request.StakeholderID, userID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, domain.NewNotFoundError(err)
			}
			return nil, domain.NewInternalServerError(err)
		}
		if name == "" {
			name = stakeholder.Name
		}
		if email == nil {
			email = stakeholder.Email
		}
		if phone == nil {
			phone = stakeholder.Phone
		}
	}

	email = normalizeConfirmerEmail(email)
	phone = normalizeConfirmerPhone(phone)

	if name == "" {
		return nil, domain.NewBadRequestError(errors.New("confirmer name is required"))
	}
	if email == nil && phone == nil {
		return nil, domain.NewBadRequestError(errors.New("either email or phone must be provided"))
	}

	existing, err := u.repository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	for _, item := range existing {
		if item.Status == domain.UserTransitionConfirmerStatusRevoked ||
			item.Status == domain.UserTransitionConfirmerStatusDeclined ||
			item.Status == domain.UserTransitionConfirmerStatusExpired {
			continue
		}
		if request.StakeholderID != nil && item.StakeholderID != nil && *item.StakeholderID == *request.StakeholderID {
			return nil, domain.NewBadRequestError(errors.New("confirmer already invited"))
		}
		if email != nil && item.Email != nil && strings.EqualFold(*item.Email, *email) {
			return nil, domain.NewBadRequestError(errors.New("confirmer email already invited"))
		}
		if phone != nil && item.Phone != nil && *item.Phone == *phone {
			return nil, domain.NewBadRequestError(errors.New("confirmer phone already invited"))
		}
	}

	token, err := securetoken.Generate()
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	now := time.Now()
	expiresAt := now.Add(confirmerInvitationTTL)

	created, err := u.repository.Create(ctx, &domain.UserTransitionConfirmer{
		StakeholderID:   request.StakeholderID,
		Name:            name,
		Email:           email,
		Phone:           phone,
		Status:          domain.UserTransitionConfirmerStatusPending,
		InvitedAt:       &now,
		InviteTokenHash: securetoken.Hash(token),
		InviteExpiresAt: &expiresAt,
	}, userID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.NewBadRequestError(errors.New("confirmer already invited"))
		}
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.TransitionInvitationResponse{
		Confirmer:       created,
		InvitationToken: token,
		ExpiresAt:       expiresAt,
	}, nil
}

func (u *userTransitionConfirmerUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.UserTransitionConfirmer, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	items, err := u.repository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	now := time.Now()
	for _, item := range items {
		if item.Status != domain.UserTransitionConfirmerStatusPending ||
			item.InviteExpiresAt == nil ||
			!now.After(*item.InviteExpiresAt) {
			continue
		}
		item.Status = domain.UserTransitionConfirmerStatusExpired
		item.InviteTokenHash = ""
		item.InviteExpiresAt = nil
		updated, updateErr := u.repository.Update(ctx, item, userID)
		if updateErr != nil {
			if errors.Is(updateErr, domain.ErrInvalidTransitionState) {
				current, findErr := u.repository.FindByID(ctx, &item.ID, userID)
				if findErr == nil {
					*item = *current
					continue
				}
			}
			return nil, domain.NewInternalServerError(updateErr)
		}
		*item = *updated
	}

	return items, nil
}

func (u *userTransitionConfirmerUseCase) Resend(c context.Context, id *domain.ID, userID *domain.ID) (*domain.TransitionInvitationResponse, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}
	if item.Status == domain.UserTransitionConfirmerStatusAccepted {
		return nil, domain.NewBadRequestError(errors.New("accepted confirmer cannot be re-invited"))
	}
	if item.Status == domain.UserTransitionConfirmerStatusRevoked {
		return nil, domain.NewBadRequestError(errors.New("revoked confirmer cannot be re-invited"))
	}

	token, err := securetoken.Generate()
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	now := time.Now()
	expiresAt := now.Add(confirmerInvitationTTL)
	item.Status = domain.UserTransitionConfirmerStatusPending
	item.InvitedAt = &now
	item.AcceptedAt = nil
	item.RevokedAt = nil
	item.InviteTokenHash = securetoken.Hash(token)
	item.InviteExpiresAt = &expiresAt
	item.AuthTokenHash = ""
	item.AuthTokenExpiresAt = nil

	updated, err := u.repository.Update(ctx, item, userID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.NewBadRequestError(errors.New("confirmer with the same contact already exists"))
		}
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return nil, domain.NewBadRequestError(errors.New("confirmer cannot be re-invited in its current state"))
		}
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.TransitionInvitationResponse{
		Confirmer:       updated,
		InvitationToken: token,
		ExpiresAt:       expiresAt,
	}, nil
}

func (u *userTransitionConfirmerUseCase) Revoke(c context.Context, id *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.repository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}
	if item.Status == domain.UserTransitionConfirmerStatusRevoked {
		return nil
	}

	now := time.Now()
	item.Status = domain.UserTransitionConfirmerStatusRevoked
	item.RevokedAt = &now
	item.InviteTokenHash = ""
	item.InviteExpiresAt = nil
	item.AuthTokenHash = ""
	item.AuthTokenExpiresAt = nil
	if _, err := u.repository.Update(ctx, item, userID); err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return domain.NewBadRequestError(errors.New("confirmer cannot be revoked in its current state"))
		}
		return domain.NewInternalServerError(err)
	}
	return nil
}

func (u *userTransitionConfirmerUseCase) Accept(c context.Context, token string) (*domain.AcceptTransitionInvitationResponse, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.findByInviteToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if item.Status != domain.UserTransitionConfirmerStatusPending {
		return nil, domain.NewBadRequestError(errors.New("invitation is no longer active"))
	}
	now := time.Now()
	if item.InviteExpiresAt == nil || !now.Before(*item.InviteExpiresAt) {
		item.Status = domain.UserTransitionConfirmerStatusExpired
		item.InviteTokenHash = ""
		item.InviteExpiresAt = nil
		_, _ = u.repository.Update(ctx, item, &item.UserID)
		return nil, domain.NewBadRequestError(domain.ErrInvitationExpired)
	}

	authToken, err := securetoken.Generate()
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	authExpiresAt := now.Add(confirmerAuthTTL)
	item.Status = domain.UserTransitionConfirmerStatusAccepted
	item.AcceptedAt = &now
	item.InviteTokenHash = ""
	item.InviteExpiresAt = nil
	item.AuthTokenHash = securetoken.Hash(authToken)
	item.AuthTokenExpiresAt = &authExpiresAt

	updated, err := u.repository.Update(ctx, item, &item.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return nil, domain.NewBadRequestError(errors.New("invitation is no longer active"))
		}
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.AcceptTransitionInvitationResponse{
		Confirmer: updated,
		AuthToken: authToken,
		ExpiresAt: authExpiresAt,
	}, nil
}

func (u *userTransitionConfirmerUseCase) Decline(c context.Context, token string) error {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	item, err := u.findByInviteToken(ctx, token)
	if err != nil {
		return err
	}
	if item.Status != domain.UserTransitionConfirmerStatusPending {
		return domain.NewBadRequestError(errors.New("invitation is no longer active"))
	}

	item.Status = domain.UserTransitionConfirmerStatusDeclined
	item.InviteTokenHash = ""
	item.InviteExpiresAt = nil
	if _, err := u.repository.Update(ctx, item, &item.UserID); err != nil {
		if errors.Is(err, domain.ErrInvalidTransitionState) {
			return domain.NewBadRequestError(errors.New("invitation is no longer active"))
		}
		return domain.NewInternalServerError(err)
	}
	return nil
}

func (u *userTransitionConfirmerUseCase) Authenticate(c context.Context, token string) (*domain.UserTransitionConfirmer, error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	if token == "" {
		return nil, domain.NewUnauthorizedError(errors.New("confirmer token is required"))
	}
	item, err := u.repository.FindByAuthTokenHash(ctx, securetoken.Hash(token))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewUnauthorizedError(errors.New("invalid confirmer token"))
		}
		return nil, domain.NewInternalServerError(err)
	}
	if item.Status != domain.UserTransitionConfirmerStatusAccepted {
		return nil, domain.NewUnauthorizedError(errors.New("confirmer is not accepted"))
	}
	if item.AuthTokenExpiresAt == nil || !time.Now().Before(*item.AuthTokenExpiresAt) {
		return nil, domain.NewUnauthorizedError(errors.New("confirmer token expired"))
	}
	return item, nil
}

func normalizeConfirmerEmail(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.ToLower(strings.TrimSpace(*value))
	if normalized == "" {
		return nil
	}
	return &normalized
}

func normalizeConfirmerPhone(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func (u *userTransitionConfirmerUseCase) findByInviteToken(c context.Context, token string) (*domain.UserTransitionConfirmer, error) {
	if token == "" {
		return nil, domain.NewBadRequestError(errors.New("invitation token is required"))
	}
	item, err := u.repository.FindByInviteTokenHash(c, securetoken.Hash(token))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(errors.New("invitation not found"))
		}
		return nil, domain.NewInternalServerError(err)
	}
	return item, nil
}
