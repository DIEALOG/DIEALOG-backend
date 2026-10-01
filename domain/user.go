package domain

import (
	"context"
	"time"
)

type User struct {
	ID        ID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

type UserRepository interface {
	Create(c context.Context, user *User) (*User, error)
	FindAll(c context.Context) ([]*User, error)
	FindByEmail(c context.Context, email string) (*User, error)
	FindByID(c context.Context, id *ID) (*User, error)
}
