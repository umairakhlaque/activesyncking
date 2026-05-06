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

	if err := run(cfg, log); err != nil {
		log.Fatal("fatal error", zap.Error(err))
	}
}

func run(cfg *config.Config, log *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := store.Open(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	// Placeholder — admin portal API handlers wired in Phase 2
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"adminsvc"}`)
	})

	srv := &http.Server{
		Addr:         cfg.Admin.Listen,
		Handler:      mux,
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
