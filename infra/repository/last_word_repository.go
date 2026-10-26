package repository

import (
	"context"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/DIEALOG/DIEALOG-backend/ent/lastword"
	"github.com/DIEALOG/DIEALOG-backend/ent/user"
	"github.com/janghanul090801/collections"
)

type lastWordRepository struct {
	client *ent.Client
}

func NewLastWordRepository(client *ent.Client) domain.LastWordRepository {
	return &lastWordRepository{
		client: client,
	}
}

func (r *lastWordRepository) Create(c context.Context, lastWord *domain.LastWord) (*domain.LastWord, error) {
	receiverIDs := collections.ToIterable(lastWord.Receiver).Map(func(s domain.Stakeholder) domain.ID {
		return s.ID
	})

	receiverGroupIDs := collections.ToIterable(lastWord.ReceiverGroups).Map(func(s domain.StakeholderGroup) domain.ID {
		return s.ID
	})

	l, err := r.client.LastWord.Create().
		SetTitle(lastWord.Title).
		SetMessage(lastWord.Message).
		SetOwnerID(lastWord.UserID).
		AddReceiverIDs(receiverIDs...).
		AddReceiverGroupIDs(receiverGroupIDs...).
		Save(c)
	if err != nil {
		return nil, err
	}

	loaded, err := r.client.LastWord.Query().
		Where(
			lastword.IDEQ(l.ID),
		).
		WithReceiver().
		WithReceiverGroups().
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainLastWord(loaded), nil
}

func (r *lastWordRepository) FindAllByUserID(c context.Context, userID *domain.ID) ([]*domain.LastWord, error) {
	lastWords, err := r.client.LastWord.Query().
		Where(
			lastword.HasOwnerWith(
				user.IDEQ(*userID),
			),
		).
		WithReceiver().
		WithOwner().
		WithReceiverGroups().
		All(c)
	if err != nil {
		return nil, err
	}

	return collections.ToIterable(lastWords).Map(toDomainLastWord), nil
}

func (r *lastWordRepository) FindByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.LastWord, error) {
	lastWord, err := r.client.LastWord.Query().
		Where(
			lastword.IDEQ(*id),
			lastword.HasOwnerWith(
				user.IDEQ(*userID),
			),
		).
		WithReceiverGroups().
		WithReceiver().
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainLastWord(lastWord), nil
}

func (r *lastWordRepository) Update(c context.Context, id *domain.ID, lastWord *domain.LastWord, userID *domain.ID) (*domain.LastWord, error) {
	receiverIDs := collections.ToIterable(lastWord.Receiver).Map(func(s domain.Stakeholder) domain.ID {
		return s.ID
	})

	receiverGroupIDs := collections.ToIterable(lastWord.ReceiverGroups).Map(func(s domain.StakeholderGroup) domain.ID {
		return s.ID
	})

	n, err := r.client.LastWord.Update().
		Where(
			lastword.IDEQ(*id),
			lastword.HasOwnerWith(user.IDEQ(*userID)),
		).
		SetTitle(lastWord.Title).
		SetMessage(lastWord.Message).
		ClearReceiver().
		AddReceiverIDs(receiverIDs...).
		ClearReceiverGroups().
		AddReceiverGroupIDs(receiverGroupIDs...).
		Save(c)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, &ent.NotFoundError{}
	}

	loaded, err := r.client.LastWord.Query().
		Where(
			lastword.IDEQ(*id),
			lastword.HasOwnerWith(user.IDEQ(*userID)),
		).
		WithReceiver().
		WithReceiverGroups().
		WithOwner().
		Only(c)
	if err != nil {
		return nil, err
	}

	return toDomainLastWord(loaded), nil
}

func (r *lastWordRepository) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	n, err := r.client.LastWord.Delete().
		Where(
			lastword.IDEQ(*id),
			lastword.HasOwnerWith(user.IDEQ(*userID)),
		).
		Exec(c)
	if err != nil {
		return err
	}
	if n == 0 {
		return &ent.NotFoundError{}
	}

	return nil
}
