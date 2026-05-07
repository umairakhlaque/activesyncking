package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

// Store holds admin-specific read/write queries that span multiple tables.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// DashboardStats holds aggregated counts for the overview page.
type DashboardStats struct {
	TotalUsers         int `json:"totalUsers"`
	ActiveUsers        int `json:"activeUsers"`
	TotalDevices       int `json:"totalDevices"`
	PendingApprovals   int `json:"pendingApprovals"`
	BlockedDevices     int `json:"blockedDevices"`
	AuthEventsLast24h  int `json:"authEventsLast24h"`
	MFAFailuresLast24h int `json:"mfaFailuresLast24h"`
}

func (s *Store) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	const q = `
		SELECT
			(SELECT COUNT(*) FROM users)::int,
			(SELECT COUNT(*) FROM users WHERE status = 'active')::int,
			(SELECT COUNT(*) FROM devices)::int,
			(SELECT COUNT(*) FROM devices WHERE status = 'pending')::int,
			(SELECT COUNT(*) FROM devices WHERE status IN ('blocked', 'quarantine'))::int,
			(SELECT COUNT(*) FROM audit_events WHERE occurred_at > NOW() - INTERVAL '24 hours')::int,
			(SELECT COUNT(*) FROM audit_events
			   WHERE occurred_at > NOW() - INTERVAL '24 hours'
			     AND action = 'MFA_FAILURE')::int`

	stats := &DashboardStats{}
	err := s.pool.QueryRow(ctx, q).Scan(
		&stats.TotalUsers, &stats.ActiveUsers,
		&stats.TotalDevices, &stats.PendingApprovals, &stats.BlockedDevices,
		&stats.AuthEventsLast24h, &stats.MFAFailuresLast24h,
	)
	if err != nil {
		return nil, fmt.Errorf("querying dashboard stats: %w", err)
	}
	return stats, nil
}

// UserRow extends models.User with admin-facing computed fields.
type UserRow struct {
	models.User
	MFAEnrolled bool `json:"mfaEnrolled"`
	DeviceCount int  `json:"deviceCount"`
}

func (s *Store) ListUsers(ctx context.Context, page, pageSize int) ([]*UserRow, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	const q = `
		SELECT u.id, u.username, u.email, u.display_name, u.status, u.source,
		       u.failed_logins, u.locked_until, u.created_at, u.updated_at,
		       COALESCE(ts.enrolled, false) AS mfa_enrolled,
		       COALESCE(dc.cnt, 0)          AS device_count
		FROM users u
		LEFT JOIN totp_secrets ts ON ts.user_id = u.id
		LEFT JOIN (
		    SELECT user_id, COUNT(*) AS cnt FROM devices GROUP BY user_id
		) dc ON dc.user_id = u.id
		ORDER BY u.username
		LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []*UserRow
	for rows.Next() {
		r := &UserRow{}
		if err := rows.Scan(
			&r.ID, &r.Username, &r.Email, &r.DisplayName, &r.Status, &r.Source,
			&r.FailedLogins, &r.LockedUntil, &r.CreatedAt, &r.UpdatedAt,
			&r.MFAEnrolled, &r.DeviceCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, r)
	}
	return users, total, rows.Err()
}

func (s *Store) SetUserStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET status = $2 WHERE id = $1`, id, status)
	return err
}

// DeviceRow extends models.Device with the owner's username.
type DeviceRow struct {
	models.Device
	Username string
}

func (s *Store) ListDevices(ctx context.Context, page, pageSize int) ([]*DeviceRow, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM devices`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting devices: %w", err)
	}

	const q = `
		SELECT d.id, d.device_eas_id, d.user_id, d.device_type, d.device_model, d.user_agent,
		       d.last_ip, d.status, d.trust_token, d.trusted_at, d.trust_expiry,
		       d.enrolled_at, d.blocked_at, d.block_reason, d.created_at, d.updated_at,
		       u.username
		FROM devices d
		JOIN users u ON u.id = d.user_id
		ORDER BY d.updated_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing devices: %w", err)
	}
	defer rows.Close()

	var devices []*DeviceRow
	for rows.Next() {
		r := &DeviceRow{}
		if err := rows.Scan(
			&r.ID, &r.DeviceEASID, &r.UserID, &r.DeviceType, &r.DeviceModel, &r.UserAgent,
			&r.LastIP, &r.Status, &r.TrustToken, &r.TrustedAt, &r.TrustExpiry,
			&r.EnrolledAt, &r.BlockedAt, &r.BlockReason, &r.CreatedAt, &r.UpdatedAt,
			&r.Username,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning device row: %w", err)
		}
		devices = append(devices, r)
	}
	return devices, total, rows.Err()
}

func (s *Store) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	return err
}

// AuditRow is a flat audit_events record for the admin API.
type AuditRow struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	Action      models.AuditAction
	Username    string
	DeviceEASID string
	ClientIP    string
	Result      string
	Severity    string // normalised: "warn" → "warning"
}

func (s *Store) ListAudit(ctx context.Context, page, pageSize int) ([]*AuditRow, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_events`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting audit events: %w", err)
	}

	const q = `
		SELECT id, occurred_at, action,
		       COALESCE(username, ''),
		       COALESCE(device_eas_id, ''),
		       COALESCE(client_ip::text, ''),
		       COALESCE(result, ''),
		       severity
		FROM audit_events
		ORDER BY occurred_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit events: %w", err)
	}
	defer rows.Close()

	var events []*AuditRow
	for rows.Next() {
		r := &AuditRow{}
		var sev models.AuditSeverity
		if err := rows.Scan(
			&r.ID, &r.OccurredAt, &r.Action,
			&r.Username, &r.DeviceEASID, &r.ClientIP, &r.Result, &sev,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning audit row: %w", err)
		}
		r.Severity = normSeverity(sev)
		events = append(events, r)
	}
	return events, total, rows.Err()
}

// normSeverity maps DB severity values to the frontend-expected string.
func normSeverity(s models.AuditSeverity) string {
	if s == models.AuditSeverityWarn {
		return "warning"
	}
	return string(s)
}
