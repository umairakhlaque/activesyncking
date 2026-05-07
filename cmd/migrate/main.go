// migrate applies SQL migrations to the SyncGuard database.
// It is run as a Fly.io release_command before each new deployment,
// ensuring the schema is always up to date before traffic shifts to the new version.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsDir = "/migrations"

func main() {
	slog.Info("SyncGuard migration runner starting")

	// Read only what we need: the database URL. We intentionally do not use
	// config.Load() here because that path requires auth secrets (JWT, encryption
	// key) that are irrelevant to schema migrations.
	dbURL := os.Getenv("SYNCGUARD_DB_DATABASE_URL")
	if dbURL == "" {
		// Fallback: build DSN from individual components
		host := envOr("SYNCGUARD_DB_HOST", "localhost")
		port := envOr("SYNCGUARD_DB_PORT", "5432")
		name := envOr("SYNCGUARD_DB_NAME", "syncguard")
		user := envOr("SYNCGUARD_DB_USER", "syncguard")
		pass := envOr("SYNCGUARD_DB_PASSWORD", "")
		ssl := envOr("SYNCGUARD_DB_SSLMODE", "disable")
		dbURL = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			host, port, name, user, pass, ssl)
	} else {
		// Append pool sizing to the URL
		sep := "?"
		if strings.Contains(dbURL, "?") {
			sep = "&"
		}
		dbURL = dbURL + sep + "pool_max_conns=2&pool_min_conns=1"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		slog.Error("invalid database URL", "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("database ping failed", "err", err)
		os.Exit(1)
	}

	if err := ensureMigrationsTable(ctx, pool); err != nil {
		slog.Error("failed to create migrations table", "err", err)
		os.Exit(1)
	}

	applied, err := appliedMigrations(ctx, pool)
	if err != nil {
		slog.Error("failed to read applied migrations", "err", err)
		os.Exit(1)
	}

	files, err := pendingMigrations(applied)
	if err != nil {
		slog.Error("failed to find migration files", "err", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		slog.Info("no pending migrations")
		return
	}

	for _, f := range files {
		if err := applyMigration(ctx, pool, f); err != nil {
			slog.Error("migration failed", "file", f, "err", err)
			os.Exit(1)
		}
		slog.Info("migration applied", "file", filepath.Base(f))
	}

	slog.Info("all migrations applied successfully", "count", len(files))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	return err
}

func appliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT filename FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func pendingMigrations(applied map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations dir %s: %w", migrationsDir, err)
	}

	var pending []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		if !applied[e.Name()] {
			pending = append(pending, filepath.Join(migrationsDir, e.Name()))
		}
	}

	sort.Strings(pending)
	return pending, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, path string) error {
	sql, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	// Use the simple query protocol so multi-statement SQL files with their
	// own BEGIN/COMMIT are executed correctly in a single round-trip.
	mrr := conn.Conn().PgConn().Exec(ctx, string(sql))
	if err := mrr.Close(); err != nil {
		return fmt.Errorf("executing migration %s: %w", filepath.Base(path), err)
	}

	if _, err := conn.Exec(ctx,
		`INSERT INTO schema_migrations (filename) VALUES ($1)`,
		filepath.Base(path),
	); err != nil {
		return fmt.Errorf("recording migration %s: %w", filepath.Base(path), err)
	}

	return nil
}
