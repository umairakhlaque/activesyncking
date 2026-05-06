package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/umairakhlaque/activesyncking/internal/auth"
	"github.com/umairakhlaque/activesyncking/internal/config"
	sgcrypto "github.com/umairakhlaque/activesyncking/internal/crypto"
	"github.com/umairakhlaque/activesyncking/internal/mfa"
	mfaemail "github.com/umairakhlaque/activesyncking/internal/mfa/email"
	mfatotp "github.com/umairakhlaque/activesyncking/internal/mfa/totp"
	"github.com/umairakhlaque/activesyncking/internal/session"
	"github.com/umairakhlaque/activesyncking/internal/store"
)

var version = "dev"

func main() {
	cfgFile := flag.String("config", "", "path to config YAML file (optional; env vars take precedence)")
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	log.Info("SyncGuard Auth Service starting", zap.String("version", version))

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	if err := run(cfg, log); err != nil {
		log.Fatal("fatal error", zap.Error(err))
	}
}

func run(cfg *config.Config, log *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── Database ──────────────────────────────────────────────────────────────
	pool, err := store.Open(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer pool.Close()
	log.Info("database connected", zap.String("host", cfg.DB.Host))

	// ── Redis ─────────────────────────────────────────────────────────────────
	sessions, err := session.NewStore(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer sessions.Close()
	log.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	// ── Encryption ────────────────────────────────────────────────────────────
	enc, err := sgcrypto.NewEncryptor(cfg.Encryption.Key)
	if err != nil {
		return fmt.Errorf("initialising encryptor: %w", err)
	}

	// ── Stores ────────────────────────────────────────────────────────────────
	users := store.NewUserStore(pool)
	devices := store.NewDeviceStore(pool)
	totpSecrets := store.NewTOTPStore(pool)
	challengeDB := store.NewChallengeDBStore(pool)
	auditStore := store.NewAuditStore(pool)

	// ── MFA Providers ─────────────────────────────────────────────────────────
	providers := []mfa.Provider{
		mfatotp.New(),
		mfaemail.New(cfg.Email),
	}

	// ── Auth Service ──────────────────────────────────────────────────────────
	svc := auth.NewService(
		cfg.Auth,
		users, devices, totpSecrets,
		challengeDB, auditStore,
		sessions,
		providers,
		enc,
		log,
	)
	handler := auth.NewHandler(svc, log)
	router := auth.NewRouter(handler)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Auth.Listen,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("auth service listening", zap.String("addr", cfg.Auth.Listen))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutting down auth service")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
	}

	return nil
}
