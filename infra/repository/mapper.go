package repository

import (
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
)

func toDomainUser(entity *ent.User) *domain.User {
	return &domain.User{
		ID:        entity.ID,
		Name:      entity.Name,
		Email:     entity.Email,
		Password:  entity.Password,
		CreatedAt: entity.CreatedAt,
	}
}

func toDomainTask(entity *ent.Task) *domain.Task {
	return &domain.Task{
		ID:        entity.ID,
		Title:     entity.Title,
		UserID:    entity.Edges.Owner.ID,
		CreatedAt: entity.CreatedAt,
	}
}
