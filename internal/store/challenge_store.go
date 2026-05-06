package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

// ChallengeDBStore persists challenge records to PostgreSQL for audit.
// The hot path reads/writes challenges from Redis (see session package).
type ChallengeDBStore struct {
	pool *pgxpool.Pool
}

func NewChallengeDBStore(pool *pgxpool.Pool) *ChallengeDBStore {
	return &ChallengeDBStore{pool: pool}
}

func (s *ChallengeDBStore) Create(ctx context.Context, c *models.Challenge) error {
	const q = `
		INSERT INTO mfa_challenges
		  (id, user_id, device_eas_id, method, status, attempts, otp_hash, client_ip, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::inet, $9, $10)`

	_, err := s.pool.Exec(ctx, q,
		c.ID, c.UserID, c.DeviceEASID, c.Method, c.Status, c.Attempts,
		c.OTPHash, c.ClientIP, c.CreatedAt, c.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating challenge record: %w", err)
	}
	return nil
}

func (s *ChallengeDBStore) UpdateStatus(ctx context.Context, id string, status models.ChallengeStatus, attempts int) error {
	const q = `
		UPDATE mfa_challenges
		SET status       = $2,
		    attempts     = $3,
		    completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END
		WHERE id = $1`

	_, err := s.pool.Exec(ctx, q, id, status, attempts)
	return err
}

func (s *ChallengeDBStore) Get(ctx context.Context, id string) (*models.Challenge, error) {
	const q = `
		SELECT id, user_id, device_eas_id, method, status, attempts, otp_hash, client_ip, created_at, expires_at
		FROM mfa_challenges
		WHERE id = $1`

	c := &models.Challenge{}
	var clientIP string
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&c.ID, &c.UserID, &c.DeviceEASID, &c.Method, &c.Status, &c.Attempts,
		&c.OTPHash, &clientIP, &c.CreatedAt, &c.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying challenge: %w", err)
	}
	c.ClientIP = clientIP

	// Mark expired challenges on read
	if c.Status == models.ChallengeStatusPending && time.Now().After(c.ExpiresAt) {
		c.Status = models.ChallengeStatusExpired
	}
	return c, nil
}
