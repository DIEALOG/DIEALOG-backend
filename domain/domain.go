package domain

import "github.com/google/uuid"

type ID = uuid.UUID

func NewID() ID {
	return uuid.New()
}

// StringToID convert a string to an entity ID
func StringToID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	return id, err
}
