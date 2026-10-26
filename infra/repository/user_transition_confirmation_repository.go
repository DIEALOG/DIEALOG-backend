package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitioncase"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitionconfirmation"
	"github.com/DIEALOG/DIEALOG-backend/ent/usertransitionconfirmer"
)

type userTransitionConfirmationRepository struct {
	client *ent.Client
}

func NewUserTransitionConfirmationRepository(client *ent.Client) domain.UserTransitionConfirmationRepository {
	return &userTransitionConfirmationRepository{client: client}
}

func (r *userTransitionConfirmationRepository) Upsert(c context.Context, confirmation *domain.UserTransitionConfirmation, userID *domain.ID) (*domain.UserTransitionConfirmation, error) {
	tx, err := r.client.Tx(c)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	casePending, err := tx.UserTransitionCase.Query().
		Where(
			usertransitioncase.IDEQ(confirmation.TransitionCaseID),
			usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			usertransitioncase.StatusEQ(string(domain.UserTransitionCaseStatusPending)),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !casePending {
		exists, queryErr := tx.UserTransitionCase.Query().
			Where(
				usertransitioncase.IDEQ(confirmation.TransitionCaseID),
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			).
			Exist(c)
		if queryErr != nil {
			return nil, queryErr
		}
		if !exists {
			return nil, &ent.NotFoundError{}
		}
		return nil, domain.ErrTransitionCaseNotEligible
	}

	confirmerAccepted, err := tx.UserTransitionConfirmer.Query().
		Where(
			usertransitionconfirmer.IDEQ(confirmation.TransitionConfirmerID),
			usertransitionconfirmer.HasOwnerWith(user.IDEQ(*userID)),
			usertransitionconfirmer.StatusEQ(string(domain.UserTransitionConfirmerStatusAccepted)),
		).
		Exist(c)
	if err != nil {
		return nil, err
	}
	if !confirmerAccepted {
		exists, queryErr := tx.UserTransitionConfirmer.Query().
			Where(
				usertransitionconfirmer.IDEQ(confirmation.TransitionConfirmerID),
				usertransitionconfirmer.HasOwnerWith(user.IDEQ(*userID)),
			).
			Exist(c)
		if queryErr != nil {
			return nil, queryErr
		}
		if !exists {
			return nil, &ent.NotFoundError{}
		}
		return nil, domain.ErrConfirmerNotAccepted
	}

	existing, err := tx.UserTransitionConfirmation.Query().
		Where(
			usertransitionconfirmation.HasTransitionCaseWith(usertransitioncase.IDEQ(confirmation.TransitionCaseID)),
			usertransitionconfirmation.HasTransitionConfirmerWith(usertransitionconfirmer.IDEQ(confirmation.TransitionConfirmerID)),
		).
		Only(c)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	var entity *ent.UserTransitionConfirmation
	if ent.IsNotFound(err) {
		entity, err = tx.UserTransitionConfirmation.Create().
			SetDecision(string(confirmation.Decision)).
			SetNillableComment(confirmation.Comment).
			SetNillableRespondedAt(confirmation.RespondedAt).
			SetTransitionCaseID(confirmation.TransitionCaseID).
			SetTransitionConfirmerID(confirmation.TransitionConfirmerID).
			Save(c)
	} else {
		update := tx.UserTransitionConfirmation.UpdateOneID(existing.ID).
			SetDecision(string(confirmation.Decision))
		if confirmation.Comment != nil {
			update.SetComment(*confirmation.Comment)
		} else {
			update.ClearComment()
		}
		if confirmation.RespondedAt != nil {
			update.SetRespondedAt(*confirmation.RespondedAt)
		} else {
			update.ClearRespondedAt()
		}
		entity, err = update.Save(c)
	}
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	item := toDomainUserTransitionConfirmation(entity)
	item.TransitionCaseID = confirmation.TransitionCaseID
	item.TransitionConfirmerID = confirmation.TransitionConfirmerID
	return item, nil
}

func (r *userTransitionConfirmationRepository) FindByCaseID(c context.Context, caseID *domain.ID, userID *domain.ID) ([]*domain.UserTransitionConfirmation, error) {
	entities, err := r.client.UserTransitionConfirmation.Query().
		Where(
			usertransitionconfirmation.HasTransitionCaseWith(
				usertransitioncase.IDEQ(*caseID),
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
		).
		WithTransitionCase().
		WithTransitionConfirmer().
		Order(ent.Asc(usertransitionconfirmation.FieldCreatedAt)).
		All(c)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.UserTransitionConfirmation, len(entities))
	for i, entity := range entities {
		items[i] = toDomainUserTransitionConfirmation(entity)
	}
	return items, nil
}

func (r *userTransitionConfirmationRepository) CountAgreeByCaseID(c context.Context, caseID *domain.ID, userID *domain.ID) (int, error) {
	return r.client.UserTransitionConfirmation.Query().
		Where(
			usertransitionconfirmation.HasTransitionCaseWith(
				usertransitioncase.IDEQ(*caseID),
				usertransitioncase.HasOwnerWith(user.IDEQ(*userID)),
			),
			usertransitionconfirmation.DecisionEQ(string(domain.UserTransitionDecisionAgree)),
			usertransitionconfirmation.HasTransitionConfirmerWith(
				usertransitionconfirmer.StatusEQ(string(domain.UserTransitionConfirmerStatusAccepted)),
			),
		).
		Count(c)
}
