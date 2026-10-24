package domain

import "context"

type Profile struct {
	Name              string             `json:"name"`
	Email             string             `json:"email"`
	PhoneNumber       string             `json:"phone_number"`
	Stakeholders      []Stakeholder      `json:"stakeholders,omitempty"`
	StakeholderGroups []StakeholderGroup `json:"stakeholder_groups,omitempty"`
}

type ProfileUseCase interface {
	GetProfileByID(c context.Context, userID *ID) (*Profile, error)
}
