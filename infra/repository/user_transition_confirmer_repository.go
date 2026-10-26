package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitionconfirmer"
)

type userTransitionConfirmerRepository struct {
	client *ent.Client
}

func NewUserTransitionConfirmerRepository(client *ent.Client) domain.UserTransitionConfirmerRepository {
	return &userTransitionConfirmerRepository{client: client}
}

func (r *userTransitionConfirmerRepository) Create(c context.Context, confirmer *domain.UserTransitionConfirmer, userID *domain.ID) (*domain.UserTransitionConfirmer, error) {
	builder := r.client.UserTransitionConfirmer.Create().
		SetName(confirmer.Name).
		SetStatus(string(confirmer.Status)).
		SetOwnerID(*userID).
		SetNillableEmail(confirmer.Email).
		SetNillablePhone(confirmer.Phone).
		SetNillableInvitedAt(confirmer.InvitedAt).
		SetNillableAcceptedAt(confirmer.AcceptedAt).
		SetNillableRevokedAt(confirmer.RevokedAt).
		SetNillableInviteExpiresAt(confirmer.InviteExpiresAt).
		SetNillableAuthTokenExpiresAt(confirmer.AuthTokenExpiresAt)

	if confirmer.StakeholderID != nil {
		builder.SetStakeholderID(*confirmer.StakeholderID)
	}
	if confirmer.InviteTokenHash != "" {
		builder.SetInviteTokenHash(confirmer.InviteTokenHash)
	}
	if confirmer.AuthTokenHash != "" {
		builder.SetAuthTokenHash(confirmer.AuthTokenHash)
	}

	entity, err := builder.Save(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionConfirmer(entity)
	item.UserID = *userID
	item.StakeholderID = confirmer.StakeholderID
	return item, nil
}

func (r *userTransitionConfirmerRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.UserTransitionConfirmer, error) {
	entities, err := r.client.UserTransitionConfirmer.Query().
		Where(usertransitionconfirmer.HasOwnerWith(user.IDEQ(*userID))).
		WithOwner().
		WithStakeholder().
		Order(ent.Asc(usertransitionconfirmer.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.UserTransitionConfirmer, len(entities))
	for i, entity := range entities {
		items[i] = toDomainUserTransitionConfirmer(entity)
		items[i].UserID = *userID
	}
	return items, nil
}

func (r *userTransitionConfirmerRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.UserTransitionConfirmer, error) {
	entity, err := r.client.UserTransitionConfirmer.Query().
		Where(
			usertransitionconfirmer.IDEQ(*id),
			usertransitionconfirmer.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithOwner().
		WithStakeholder().
		Only(c)
	if err != nil {
		return nil, err
	}

	item := toDomainUserTransitionConfirmer(entity)
	item.UserID = *userID
	return item, nil
}

func (r *userTransitionConfirmerRepository) FindByInviteTokenHash(c context.Context, tokenHash string) (*domain.UserTransitionConfirmer, error) {
	entity, err := r.client.UserTransitionConfirmer.Query().
		Where(usertransitionconfirmer.InviteTokenHashEQ(tokenHash)).
		WithOwner().
		WithStakeholder().
		Only(c)
	if err != nil {
		return nil, err
	}
	return toDomainUserTransitionConfirmer(entity), nil
}

func (r *userTransitionConfirmerRepository) FindByAuthTokenHash(c context.Context, tokenHash string) (*domain.UserTransitionConfirmer, error) {
	entity, err := r.client.UserTransitionConfirmer.Query().
		Where(usertransitionconfirmer.AuthTokenHashEQ(tokenHash)).
		WithOwner().
		WithStakeholder().
		Only(c)
	if err != nil {
		return nil, err
	}
	return toDomainUserTransitionConfirmer(entity), nil
}

func (r *userTransitionConfirmerRepository) Update(c context.Context, confirmer *domain.UserTransitionConfirmer, userID *domain.ID) (*domain.UserTransitionConfirmer, error) {
	builder := r.client.UserTransitionConfirmer.Update().
		Where(
			usertransitionconfirmer.IDEQ(confirmer.ID),
			usertransitionconfirmer.HasOwnerWith(user.IDEQ(*userID)),
		)

	// Guard invitation/auth status transitions so a concurrent revoke or
	// acceptance cannot be overwritten by a stale read.
	switch confirmer.Status {
	case domain.UserTransitionConfirmerStatusAccepted,
		domain.UserTransitionConfirmerStatusDeclined,
		domain.UserTransitionConfirmerStatusExpired:
		builder.Where(usertransitionconfirmer.StatusEQ(string(domain.UserTransitionConfirmerStatusPending)))
	case domain.UserTransitionConfirmerStatusPending:
		builder.Where(usertransitionconfirmer.StatusIn(
			string(domain.UserTransitionConfirmerStatusPending),
			string(domain.UserTransitionConfirmerStatusDeclined),
			string(domain.UserTransitionConfirmerStatusExpired),
		))
	case domain.UserTransitionConfirmerStatusRevoked:
		builder.Where(usertransitionconfirmer.StatusIn(
			string(domain.UserTransitionConfirmerStatusPending),
			string(domain.UserTransitionConfirmerStatusAccepted),
			string(domain.UserTransitionConfirmerStatusDeclined),
			string(domain.UserTransitionConfirmerStatusExpired),
		))
	}

	builder.
		SetName(confirmer.Name).
		SetStatus(string(confirmer.Status))

	if confirmer.Email != nil {
		builder.SetEmail(*confirmer.Email)
	} else {
		builder.ClearEmail()
	}
	if confirmer.Phone != nil {
		builder.SetPhone(*confirmer.Phone)
	} else {
		builder.ClearPhone()
	}
	if confirmer.InvitedAt != nil {
		builder.SetInvitedAt(*confirmer.InvitedAt)
	} else {
		builder.ClearInvitedAt()
	}
	if confirmer.AcceptedAt != nil {
		builder.SetAcceptedAt(*confirmer.AcceptedAt)
	} else {
		builder.ClearAcceptedAt()
	}
	if confirmer.RevokedAt != nil {
		builder.SetRevokedAt(*confirmer.RevokedAt)
	} else {
		builder.ClearRevokedAt()
	}
	if confirmer.InviteExpiresAt != nil {
		builder.SetInviteExpiresAt(*confirmer.InviteExpiresAt)
	} else {
		builder.ClearInviteExpiresAt()
	}
	if confirmer.AuthTokenExpiresAt != nil {
		builder.SetAuthTokenExpiresAt(*confirmer.AuthTokenExpiresAt)
	} else {
		builder.ClearAuthTokenExpiresAt()
	}

	if confirmer.StakeholderID != nil {
		builder.SetStakeholderID(*confirmer.StakeholderID)
	} else {
		builder.ClearStakeholder()
	}
	if confirmer.InviteTokenHash != "" {
		builder.SetInviteTokenHash(confirmer.InviteTokenHash)
	} else {
		builder.ClearInviteTokenHash()
	}
	if confirmer.AuthTokenHash != "" {
		builder.SetAuthTokenHash(confirmer.AuthTokenHash)
	} else {
		builder.ClearAuthTokenHash()
	}

	n, err := builder.Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		current, findErr := r.FindByID(c, &confirmer.ID, userID)
		if findErr != nil {
			return nil, findErr
		}
		// Acceptance issues a new auth token, so a concurrent acceptance must
		// fail instead of returning the row with a token that was never stored.
		if confirmer.Status == domain.UserTransitionConfirmerStatusAccepted &&
			current.Status == domain.UserTransitionConfirmerStatusAccepted {
			return nil, domain.ErrInvalidTransitionState
		}
		if current.Status == confirmer.Status {
			return current, nil
		}
		return nil, domain.ErrInvalidTransitionState
	}

	return r.FindByID(c, &confirmer.ID, userID)
}
