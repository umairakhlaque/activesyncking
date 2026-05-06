package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

type DeviceStore struct {
	pool *pgxpool.Pool
}

func NewDeviceStore(pool *pgxpool.Pool) *DeviceStore {
	return &DeviceStore{pool: pool}
}

func (s *DeviceStore) GetByEASIDAndUser(ctx context.Context, easID string, userID uuid.UUID) (*models.Device, error) {
	const q = `
		SELECT id, device_eas_id, user_id, device_type, device_model, user_agent,
		       last_ip, status, trust_token, trusted_at, trust_expiry,
		       enrolled_at, blocked_at, block_reason, created_at, updated_at
		FROM devices
		WHERE device_eas_id = $1 AND user_id = $2
		LIMIT 1`

	d := &models.Device{}
	err := s.pool.QueryRow(ctx, q, easID, userID).Scan(
		&d.ID, &d.DeviceEASID, &d.UserID, &d.DeviceType, &d.DeviceModel, &d.UserAgent,
		&d.LastIP, &d.Status, &d.TrustToken, &d.TrustedAt, &d.TrustExpiry,
		&d.EnrolledAt, &d.BlockedAt, &d.BlockReason, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying device: %w", err)
	}
	return d, nil
}

func (s *DeviceStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	const q = `
		SELECT id, device_eas_id, user_id, device_type, device_model, user_agent,
		       last_ip, status, trust_token, trusted_at, trust_expiry,
		       enrolled_at, blocked_at, block_reason, created_at, updated_at
		FROM devices
		WHERE id = $1`

	d := &models.Device{}
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&d.ID, &d.DeviceEASID, &d.UserID, &d.DeviceType, &d.DeviceModel, &d.UserAgent,
		&d.LastIP, &d.Status, &d.TrustToken, &d.TrustedAt, &d.TrustExpiry,
		&d.EnrolledAt, &d.BlockedAt, &d.BlockReason, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying device by id: %w", err)
	}
	return d, nil
}

// Upsert creates or updates a device record. On conflict (eas_id + user_id),
// it updates mutable metadata only — it never downgrades an approved device.
func (s *DeviceStore) Upsert(ctx context.Context, d *models.Device) error {
	const q = `
		INSERT INTO devices (id, device_eas_id, user_id, device_type, device_model, user_agent, last_ip, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7::inet, $8)
		ON CONFLICT (device_eas_id, user_id) DO UPDATE
		  SET device_type  = EXCLUDED.device_type,
		      device_model = EXCLUDED.device_model,
		      user_agent   = EXCLUDED.user_agent,
		      last_ip      = EXCLUDED.last_ip
		RETURNING id, status, created_at, updated_at`

	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	err := s.pool.QueryRow(ctx, q,
		d.ID, d.DeviceEASID, d.UserID, d.DeviceType, d.DeviceModel, d.UserAgent, d.LastIP, d.Status,
	).Scan(&d.ID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting device: %w", err)
	}
	return nil
}

func (s *DeviceStore) Approve(ctx context.Context, id uuid.UUID, trustToken string, expiry time.Time) error {
	const q = `
		UPDATE devices
		SET status       = 'approved',
		    trust_token  = $2,
		    trusted_at   = NOW(),
		    trust_expiry = $3,
		    enrolled_at  = COALESCE(enrolled_at, NOW())
		WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id, trustToken, expiry)
	return err
}

func (s *DeviceStore) SetPending(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE devices SET status = 'pending' WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id)
	return err
}

func (s *DeviceStore) Block(ctx context.Context, id uuid.UUID, reason string) error {
	const q = `
		UPDATE devices
		SET status       = 'blocked',
		    blocked_at   = NOW(),
		    block_reason = $2,
		    trust_token  = NULL
		WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id, reason)
	return err
}

func (s *DeviceStore) Quarantine(ctx context.Context, id uuid.UUID, reason string) error {
	const q = `
		UPDATE devices
		SET status       = 'quarantine',
		    blocked_at   = NOW(),
		    block_reason = $2,
		    trust_token  = NULL
		WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id, reason)
	return err
}

func (s *DeviceStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Device, error) {
	const q = `
		SELECT id, device_eas_id, user_id, device_type, device_model, user_agent,
		       last_ip, status, trust_token, trusted_at, trust_expiry,
		       enrolled_at, blocked_at, block_reason, created_at, updated_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("listing devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d := &models.Device{}
		if err := rows.Scan(
			&d.ID, &d.DeviceEASID, &d.UserID, &d.DeviceType, &d.DeviceModel, &d.UserAgent,
			&d.LastIP, &d.Status, &d.TrustToken, &d.TrustedAt, &d.TrustExpiry,
			&d.EnrolledAt, &d.BlockedAt, &d.BlockReason, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
