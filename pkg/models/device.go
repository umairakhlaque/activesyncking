package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceStatus string

const (
	DeviceStatusUnknown    DeviceStatus = "unknown"
	DeviceStatusPending    DeviceStatus = "pending"
	DeviceStatusApproved   DeviceStatus = "approved"
	DeviceStatusBlocked    DeviceStatus = "blocked"
	DeviceStatusQuarantine DeviceStatus = "quarantine"
)

type Device struct {
	ID          uuid.UUID    `db:"id"`
	DeviceEASID string       `db:"device_eas_id"` // MS-ASDevice header value
	UserID      uuid.UUID    `db:"user_id"`
	DeviceType  string       `db:"device_type"`
	DeviceModel string       `db:"device_model"`
	UserAgent   string       `db:"user_agent"`
	LastIP      string       `db:"last_ip"`
	Status      DeviceStatus `db:"status"`
	TrustToken  *string      `db:"trust_token"`
	TrustedAt   *time.Time   `db:"trusted_at"`
	TrustExpiry *time.Time   `db:"trust_expiry"`
	EnrolledAt  *time.Time   `db:"enrolled_at"`
	BlockedAt   *time.Time   `db:"blocked_at"`
	BlockReason string       `db:"block_reason"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

func (d *Device) IsTrusted() bool {
	if d.Status != DeviceStatusApproved {
		return false
	}
	if d.TrustExpiry != nil && time.Now().After(*d.TrustExpiry) {
		return false
	}
	return true
}

func (d *Device) IsBlocked() bool {
	return d.Status == DeviceStatusBlocked || d.Status == DeviceStatusQuarantine
}
