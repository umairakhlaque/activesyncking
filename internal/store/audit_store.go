package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

type AuditStore struct {
	pool *pgxpool.Pool
}

func NewAuditStore(pool *pgxpool.Pool) *AuditStore {
	return &AuditStore{pool: pool}
}

func (s *AuditStore) Write(ctx context.Context, e *models.AuditEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	detailsJSON, err := json.Marshal(e.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	const q = `
		INSERT INTO audit_events
		  (id, occurred_at, action, user_id, username, device_eas_id,
		   client_ip, challenge_id, result, details, severity)
		VALUES
		  ($1, NOW(), $2, $3, $4, $5, $6::inet, $7, $8, $9, $10)`

	_, err = s.pool.Exec(ctx, q,
		e.ID, e.Action, e.UserID, e.Username, e.DeviceEASID,
		e.ClientIP, e.ChallengeID, e.Result, detailsJSON, e.Severity,
	)
	if err != nil {
		return fmt.Errorf("writing audit event: %w", err)
	}
	return nil
}

func (s *AuditStore) ListRecent(ctx context.Context, limit int) ([]*models.AuditEvent, error) {
	const q = `
		SELECT id, occurred_at, action, user_id, username, device_eas_id,
		       client_ip, challenge_id, result, severity
		FROM audit_events
		ORDER BY occurred_at DESC
		LIMIT $1`

	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("listing audit events: %w", err)
	}
	defer rows.Close()

	var events []*models.AuditEvent
	for rows.Next() {
		e := &models.AuditEvent{}
		if err := rows.Scan(
			&e.ID, &e.OccurredAt, &e.Action, &e.UserID, &e.Username, &e.DeviceEASID,
			&e.ClientIP, &e.ChallengeID, &e.Result, &e.Severity,
		); err != nil {
			return nil, fmt.Errorf("scanning audit event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// WriteAsync fires audit writes in a goroutine so they never block the request path.
// Errors are swallowed here; callers should use Write when durability matters.
func (s *AuditStore) WriteAsync(e *models.AuditEvent) {
	go func() {
		ctx := context.Background()
		_ = s.Write(ctx, e)
	}()
}
