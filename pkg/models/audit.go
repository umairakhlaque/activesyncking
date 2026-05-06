package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string
type AuditSeverity string

const (
	AuditActionAuthAttempt     AuditAction = "AUTH_ATTEMPT"
	AuditActionAuthSuccess     AuditAction = "AUTH_SUCCESS"
	AuditActionAuthFailure     AuditAction = "AUTH_FAILURE"
	AuditActionAuthDenied      AuditAction = "AUTH_DENIED"
	AuditActionMFARequired     AuditAction = "MFA_REQUIRED"
	AuditActionMFASuccess      AuditAction = "MFA_SUCCESS"
	AuditActionMFAFailure      AuditAction = "MFA_FAILURE"
	AuditActionMFALocked       AuditAction = "MFA_LOCKED"
	AuditActionDeviceSeen      AuditAction = "DEVICE_SEEN"
	AuditActionDeviceEnrolled  AuditAction = "DEVICE_ENROLLED"
	AuditActionDeviceApproved  AuditAction = "DEVICE_APPROVED"
	AuditActionDeviceBlocked   AuditAction = "DEVICE_BLOCKED"
	AuditActionDeviceQuarantine AuditAction = "DEVICE_QUARANTINE"
	AuditActionAccessAllowed   AuditAction = "ACCESS_ALLOWED"
	AuditActionAccessDenied    AuditAction = "ACCESS_DENIED"
	AuditActionProxyForward    AuditAction = "PROXY_FORWARD"

	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarn     AuditSeverity = "warn"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

type AuditEvent struct {
	ID          uuid.UUID     `db:"id"`
	OccurredAt  time.Time     `db:"occurred_at"`
	Action      AuditAction   `db:"action"`
	UserID      *uuid.UUID    `db:"user_id"`
	Username    string        `db:"username"`
	DeviceEASID string        `db:"device_eas_id"`
	ClientIP    string        `db:"client_ip"`
	ChallengeID string        `db:"challenge_id"`
	Result      string        `db:"result"`
	Details     map[string]any `db:"-"` // serialised to JSONB
	Severity    AuditSeverity `db:"severity"`
}
