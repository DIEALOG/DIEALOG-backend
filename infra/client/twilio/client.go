package twilio

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrVerificationNotFound = errors.New("verification not found")
	ErrVerificationExpired  = errors.New("verification expired")
)

type MockPhoneVerificationClient struct {
	mu            sync.Mutex
	mockCode      string
	verifications map[string]mockVerification
}

type mockVerification struct {
	code      string
	expiresAt time.Time
	approved  bool
}

func NewMockPhoneVerificationClient() *MockPhoneVerificationClient {
	return &MockPhoneVerificationClient{
		mockCode:      "123456",
		verifications: make(map[string]mockVerification),
	}
}

func (c *MockPhoneVerificationClient) SendVerification(ctx context.Context, phoneNumber string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.verifications[phoneNumber] = mockVerification{
		code:      c.mockCode,
		expiresAt: time.Now().Add(5 * time.Minute),
	}

	return nil
}

func (c *MockPhoneVerificationClient) CheckVerification(ctx context.Context, phoneNumber string, code string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	verification, exists := c.verifications[phoneNumber]
	if !exists {
		return false, ErrVerificationNotFound
	}

	if time.Now().After(verification.expiresAt) {
		delete(c.verifications, phoneNumber)
		return false, ErrVerificationExpired
	}

	if verification.approved {
		return true, nil
	}

	if verification.code != code {
		return false, nil
	}

	verification.approved = true
	c.verifications[phoneNumber] = verification

	return true, nil
}
