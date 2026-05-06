package totp_test

import (
	"context"
	"testing"
	"time"

	otplib "github.com/pquerna/otp/totp"
	"github.com/umairakhlaque/activesyncking/internal/mfa"
	"github.com/umairakhlaque/activesyncking/internal/mfa/totp"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

func TestTOTPRoundTrip(t *testing.T) {
	p := totp.New()

	if p.Method() != models.ChallengeMethodTOTP {
		t.Fatalf("expected TOTP method, got %s", p.Method())
	}

	secret, _, err := p.GenerateSecret(context.Background(), "user-id", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateSecret: %v", err)
	}
	if secret == "" {
		t.Fatal("expected non-empty secret")
	}

	// Generate a valid OTP from the secret and validate it
	code, err := otplib.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("generating code: %v", err)
	}

	if err := p.Validate(context.Background(), code, secret, &models.Challenge{}); err != nil {
		t.Fatalf("Validate with correct code: %v", err)
	}
}

func TestTOTPInvalidCode(t *testing.T) {
	p := totp.New()
	secret, _, err := p.GenerateSecret(context.Background(), "user-id", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateSecret: %v", err)
	}

	err = p.Validate(context.Background(), "000000", secret, &models.Challenge{})
	if err == nil {
		t.Fatal("expected error for invalid OTP, got nil")
	}
	var invalidErr mfa.ErrInvalidOTP
	if _, ok := err.(mfa.ErrInvalidOTP); !ok {
		t.Fatalf("expected ErrInvalidOTP, got %T: %v", err, err)
	}
	_ = invalidErr
}
