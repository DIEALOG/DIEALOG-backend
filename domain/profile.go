package domain

import "context"

type Profile struct {
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Testament    string        `json:"testament"`
	Stakeholders []Stakeholder `json:"stakeholders,omitempty"`
}

type UpdateTestamentRequest struct {
	Testament string `form:"testament" binding:"required" json:"testament"`
}

type ProfileUseCase interface {
	GetProfileByID(c context.Context, userID *ID) (*Profile, error)
	UpdateTestament(c context.Context, userID *ID, testament string) error
}
