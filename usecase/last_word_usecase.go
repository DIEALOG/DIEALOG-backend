package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
)

type lastWordUseCase struct {
	lastWordRepository domain.LastWordRepository
	contextTimeout     time.Duration
}

func NewLastWordUseCase(lastWordRepository domain.LastWordRepository, timeout time.Duration) domain.LastWordUseCase {
	return &lastWordUseCase{
		lastWordRepository: lastWordRepository,
		contextTimeout:     timeout,
	}
}

func (l *lastWordUseCase) Create(c context.Context, lastWord *domain.LastWord) (*domain.LastWord, error) {
	if !lastWord.IsValid() {
		return nil, domain.NewBadRequestError(errors.New("title and message are required"))
	}

	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	lw, err := l.lastWordRepository.Create(ctx, lastWord)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, domain.NewBadRequestError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return lw, nil
}

func (l *lastWordUseCase) ListByUserID(c context.Context, userID *domain.ID) ([]*domain.LastWord, error) {
	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	lastWords, err := l.lastWordRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	return lastWords, nil
}

func (l *lastWordUseCase) GetByID(c context.Context, id *domain.ID, userID *domain.ID) (*domain.LastWord, error) {
	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	lw, err := l.lastWordRepository.FindByID(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return lw, nil
}

func (l *lastWordUseCase) Update(c context.Context, id *domain.ID, lastWord *domain.LastWord, userID *domain.ID) (*domain.LastWord, error) {
	if !lastWord.IsValid() {
		return nil, domain.NewBadRequestError(errors.New("title and message are required"))
	}

	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	lw, err := l.lastWordRepository.Update(ctx, id, lastWord, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		if ent.IsConstraintError(err) {
			return nil, domain.NewBadRequestError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return lw, nil
}

func (l *lastWordUseCase) Delete(c context.Context, id *domain.ID, userID *domain.ID) error {
	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	err := l.lastWordRepository.Delete(ctx, id, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.NewNotFoundError(err)
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}
