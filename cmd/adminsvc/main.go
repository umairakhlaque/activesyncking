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

	"github.com/umairakhlaque/activesyncking/internal/admin"
	"github.com/umairakhlaque/activesyncking/internal/config"
	"github.com/umairakhlaque/activesyncking/internal/store"
)

var version = "dev"

func main() {
	cfgFile := flag.String("config", "", "path to config YAML file")
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	log.Info("SyncGuard Admin Service starting", zap.String("version", version))

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}
	if cfg.Admin.APIKey == "" {
		log.Fatal("admin.api_key must not be empty — set SYNCGUARD_ADMIN_API_KEY")
	}

	if err := run(cfg, log); err != nil {
		log.Fatal("fatal error", zap.Error(err))
	}
}

func run(cfg *config.Config, log *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Use lazy open so the service starts even when Neon is in cold-start.
	// The first DB-backed request may be slow while the connection warms up.
	pool, err := store.OpenLazy(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer pool.Close()

	adminStore := admin.NewStore(pool)
	deviceStore := store.NewDeviceStore(pool)
	userStore := store.NewUserStore(pool)

	handler := admin.NewHandler(adminStore, deviceStore, userStore, cfg.Admin.APIKey, log)
	router := admin.NewRouter(handler, cfg.Admin.APIKey, cfg.Admin.CORSOrigins)

	srv := &http.Server{
		Addr:         cfg.Admin.Listen,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("admin service listening", zap.String("addr", cfg.Admin.Listen))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutting down admin service")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}

	return nil
}
