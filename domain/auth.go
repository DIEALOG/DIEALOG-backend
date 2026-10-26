package domain

import "context"

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthUseCase interface {
	Register(c context.Context, name, email, password, phoneNumber string) (*User, error)
	Login(c context.Context, email, password string) (*User, error)
	CreateAccessAndRefreshToken(c context.Context, user *User) (string, string, error)
	ExtractUserFromRefreshToken(c context.Context, requestToken string) (*User, error)
}
