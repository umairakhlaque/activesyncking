package mfa

import (
	"context"

	"github.com/umairakhlaque/activesyncking/pkg/models"
)

// Provider is the interface every MFA method must implement.
// The gateway calls Validate on the hot path; Generate is used during enrolment
// or when issuing a one-time code (email/SMS).
type Provider interface {
	// Method returns the challenge method this provider handles.
	Method() models.ChallengeMethod

	// GenerateSecret creates a new user secret (TOTP) or sends a one-time code
	// (email/SMS). For TOTP it returns the provisioning URL; for OOB methods it
	// returns an empty string (code is sent out-of-band).
	GenerateSecret(ctx context.Context, userID string, email string) (secret string, provisionURL string, err error)

	// Validate checks the submitted OTP against the stored secret or hash.
	// challenge contains the OTP hash for email/SMS; secret is the decrypted
	// TOTP base32 secret for TOTP.
	Validate(ctx context.Context, otp string, secret string, challenge *models.Challenge) error
}

// ErrInvalidOTP is returned when the OTP does not match.
type ErrInvalidOTP struct{ Msg string }

func (e ErrInvalidOTP) Error() string { return e.Msg }

// ErrExpiredOTP is returned when the OTP window has closed.
type ErrExpiredOTP struct{}

func (e ErrExpiredOTP) Error() string { return "OTP expired" }
