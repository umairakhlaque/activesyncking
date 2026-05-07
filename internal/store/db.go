package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umairakhlaque/activesyncking/internal/config"
)

// Open creates a pgxpool connection pool and verifies connectivity with a ping.
// Use OpenLazy when the service must start even if the database is temporarily
// unavailable (e.g. a cold-start serverless DB like Neon).
func Open(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	pool, err := OpenLazy(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return pool, nil
}

// OpenLazy creates a pgxpool connection pool without verifying connectivity.
// Connections are established on first use; the caller must handle transient
// errors from the first few queries while the pool warms up.
func OpenLazy(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing db dsn: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}
	return pool, nil
}
