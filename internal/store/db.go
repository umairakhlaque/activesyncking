package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umairakhlaque/activesyncking/internal/config"
)

// Open creates and validates a pgxpool connection pool.
func Open(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing db dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
