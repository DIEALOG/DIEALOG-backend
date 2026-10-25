package domain

import "context"

type CreateLastWordRequest struct {
	Title            string `json:"title"`
	Message          string `json:"message"`
	ReceiverIDs      []ID   `json:"receiver_ids"`
	ReceiverGroupIDs []ID   `json:"receiver_group_ids"`
}

type UpdateLastWordRequest struct {
	Title            string `json:"title"`
	Message          string `json:"message"`
	ReceiverIDs      []ID   `json:"receiver_ids"`
	ReceiverGroupIDs []ID   `json:"receiver_group_ids"`
}

type LastWord struct {
	ID             ID                 `json:"id"`
	UserID         ID                 `json:"user_id"`
	Receiver       []Stakeholder      `json:"receiver"`
	ReceiverGroups []StakeholderGroup `json:"receiver_groups"`
	Title          string             `json:"title"`
	Message        string             `json:"message"`
}

type LastWordRepository interface {
	Create(c context.Context, lastWord *LastWord) (*LastWord, error)
	FindAllByUserID(c context.Context, userID *ID) ([]*LastWord, error)
	FindByID(c context.Context, id *ID, userID *ID) (*LastWord, error)
	Update(c context.Context, id *ID, lastWord *LastWord, userID *ID) (*LastWord, error)
	Delete(c context.Context, id *ID, userID *ID) error
}

type LastWordUseCase interface {
	Create(c context.Context, lastWord *LastWord) (*LastWord, error)
	ListByUserID(c context.Context, userID *ID) ([]*LastWord, error)
	GetByID(c context.Context, id *ID, userID *ID) (*LastWord, error)
	Update(c context.Context, id *ID, lastWord *LastWord, userID *ID) (*LastWord, error)
	Delete(c context.Context, id *ID, userID *ID) error
}
