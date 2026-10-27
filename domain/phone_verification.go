package domain

import "context"

type SendPhoneVerificationRequest struct {
	PhoneNumber string `json:"phoneNumber"`
}

type VerifyPhoneVerificationRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Code        string `json:"code"`
}

type PhoneVerificationClient interface {
	SendVerification(c context.Context, phoneNumber string) error
	CheckVerification(c context.Context, phoneNumber string, code string) (bool, error)
}

type PhoneVerificationUseCase interface {
	SendCode(c context.Context, phoneNumber string) error
	VerifyCode(c context.Context, phoneNumber string, code string) error
}
