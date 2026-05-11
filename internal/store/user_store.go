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

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	const q = `
		SELECT id, username, email, display_name, status, source,
		       password_hash, failed_logins, locked_until, created_at, updated_at
		FROM users
		WHERE username = $1
		LIMIT 1`

	u := &models.User{}
	err := s.pool.QueryRow(ctx, q, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Status, &u.Source,
		&u.PasswordHash, &u.FailedLogins, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by username: %w", err)
	}
	return u, nil
}

func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if s.pool == nil {
		return nil, ErrNotFound
	}
	const q = `
		SELECT id, username, email, display_name, status, source,
		       password_hash, failed_logins, locked_until, created_at, updated_at
		FROM users
		WHERE id = $1`

	u := &models.User{}
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Status, &u.Source,
		&u.PasswordHash, &u.FailedLogins, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return u, nil
}

func (s *UserStore) Create(ctx context.Context, u *models.User) error {
	const q = `
		INSERT INTO users (id, username, email, display_name, status, source, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	err := s.pool.QueryRow(ctx, q,
		u.ID, u.Username, u.Email, u.DisplayName, u.Status, u.Source, u.PasswordHash,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (s *UserStore) IncrementFailedLogins(ctx context.Context, id uuid.UUID, lockUntil *time.Time) error {
	const q = `
		UPDATE users
		SET failed_logins = failed_logins + 1,
		    locked_until  = $2,
		    status        = CASE WHEN $2 IS NOT NULL THEN 'locked'::user_status ELSE status END
		WHERE id = $1`

	_, err := s.pool.Exec(ctx, q, id, lockUntil)
	return err
}

func (s *UserStore) ResetFailedLogins(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET failed_logins = 0, locked_until = NULL WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id)
	return err
}

// ─── TOTP Secrets ─────────────────────────────────────────────────────────────

type TOTPStore struct {
	pool *pgxpool.Pool
}

func NewTOTPStore(pool *pgxpool.Pool) *TOTPStore {
	return &TOTPStore{pool: pool}
}

func (s *TOTPStore) Get(ctx context.Context, userID uuid.UUID) (*models.TOTPSecret, error) {
	const q = `
		SELECT id, user_id, secret_enc, enrolled, enrolled_at, created_at, updated_at
		FROM totp_secrets
		WHERE user_id = $1`

	t := &models.TOTPSecret{}
	err := s.pool.QueryRow(ctx, q, userID).Scan(
		&t.ID, &t.UserID, &t.SecretEnc, &t.Enrolled, &t.EnrolledAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying totp secret: %w", err)
	}
	return t, nil
}

func (s *TOTPStore) Upsert(ctx context.Context, t *models.TOTPSecret) error {
	const q = `
		INSERT INTO totp_secrets (id, user_id, secret_enc, enrolled, enrolled_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		  SET secret_enc  = EXCLUDED.secret_enc,
		      enrolled    = EXCLUDED.enrolled,
		      enrolled_at = EXCLUDED.enrolled_at`

	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	_, err := s.pool.Exec(ctx, q, t.ID, t.UserID, t.SecretEnc, t.Enrolled, t.EnrolledAt)
	if err != nil {
		return fmt.Errorf("upserting totp secret: %w", err)
	}
	return nil
}

func (s *TOTPStore) MarkEnrolled(ctx context.Context, userID uuid.UUID) error {
	const q = `
		UPDATE totp_secrets
		SET enrolled = true, enrolled_at = NOW()
		WHERE user_id = $1`
	_, err := s.pool.Exec(ctx, q, userID)
	return err
}
