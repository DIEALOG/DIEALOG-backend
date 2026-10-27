package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/DIEALOG/DIEALOG-backend/domain"
)

var (
	ErrInvalidPhoneNumber      = domain.NewBadRequestError(errors.New("invalid phone number"))
	ErrInvalidVerificationCode = domain.NewBadRequestError(errors.New("invalid verification code"))
	ErrVerificationRejected    = domain.NewBadRequestError(errors.New("phone verification rejected"))
)

type phoneVerificationUseCase struct {
	client domain.PhoneVerificationClient
}

func NewPhoneVerificationUseCase(client domain.PhoneVerificationClient) domain.PhoneVerificationUseCase {
	return &phoneVerificationUseCase{
		client: client,
	}
}

var (
	e164PhoneNumberRegex  = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
	verificationCodeRegex = regexp.MustCompile(`^[0-9]{4,10}$`)
)

func (u *phoneVerificationUseCase) SendCode(c context.Context, phoneNumber string) error {
	phoneNumber = strings.TrimSpace(phoneNumber)

	if !e164PhoneNumberRegex.MatchString(phoneNumber) {
		return ErrInvalidPhoneNumber
	}

	return u.client.SendVerification(c, phoneNumber)
}

func (u *phoneVerificationUseCase) VerifyCode(c context.Context, phoneNumber string, code string) error {
	phoneNumber = strings.TrimSpace(phoneNumber)
	code = strings.TrimSpace(code)

	if !e164PhoneNumberRegex.MatchString(phoneNumber) {
		return ErrInvalidPhoneNumber
	}

	if !verificationCodeRegex.MatchString(code) {
		return ErrInvalidVerificationCode
	}

	approved, err := u.client.CheckVerification(c, phoneNumber, code)
	if err != nil {
		return err
	}

	if !approved {
		return ErrVerificationRejected
	}

	return nil
}
