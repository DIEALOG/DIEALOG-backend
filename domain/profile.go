package domain

import (
	"context"
	"mime/multipart"
)

type Profile struct {
	ID                ID                 `json:"id"`
	Name              string             `json:"name"`
	Email             string             `json:"email"`
	PhoneNumber       string             `json:"phone_number"`
	ProfileImagePath  *string            `json:"profile_image_path"`
	Stakeholders      []Stakeholder      `json:"stakeholders,omitempty"`
	StakeholderGroups []StakeholderGroup `json:"stakeholder_groups,omitempty"`
}

type ProfileUseCase interface {
	GetProfileByID(c context.Context, userID *ID) (*Profile, error)
	UpdateProfileImage(c context.Context, userID *ID, file *multipart.FileHeader) (*Profile, error)
}
