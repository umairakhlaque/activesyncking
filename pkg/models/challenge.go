package models

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeMethod string
type ChallengeStatus string

const (
	ChallengeMethodTOTP  ChallengeMethod = "totp"
	ChallengeMethodEmail ChallengeMethod = "email"
	ChallengeMethodPush  ChallengeMethod = "push"
	ChallengeMethodSMS   ChallengeMethod = "sms"

	ChallengeStatusPending   ChallengeStatus = "pending"
	ChallengeStatusCompleted ChallengeStatus = "completed"
	ChallengeStatusExpired   ChallengeStatus = "expired"
	ChallengeStatusFailed    ChallengeStatus = "failed"
	ChallengeStatusLocked    ChallengeStatus = "locked"
)

// Challenge lives primarily in Redis for sub-millisecond lookup.
// It is also written to mfa_challenges in PostgreSQL for audit completeness.
type Challenge struct {
	ID          string          `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	DeviceEASID string          `json:"device_eas_id"`
	Method      ChallengeMethod `json:"method"`
	Status      ChallengeStatus `json:"status"`
	Attempts    int             `json:"attempts"`
	OTPHash     string          `json:"otp_hash,omitempty"` // bcrypt; email/SMS only
	ClientIP    string          `json:"client_ip"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
}

func (c *Challenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *Challenge) IsTerminal() bool {
	return c.Status == ChallengeStatusCompleted ||
		c.Status == ChallengeStatusExpired ||
		c.Status == ChallengeStatusLocked
}
