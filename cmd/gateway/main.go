package main

import (
	"context"
	"crypto/tls"
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
	"github.com/umairakhlaque/activesyncking/internal/proxy"
)

var version = "dev"

func main() {
	cfgFile := flag.String("config", "", "path to config YAML file")
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	log.Info("SyncGuard EAS Gateway starting", zap.String("version", version))

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

	gw, err := proxy.New(cfg.Gateway, log)
	if err != nil {
		return fmt.Errorf("creating gateway: %w", err)
	}

	srv := &http.Server{
		Addr:         cfg.Gateway.Listen,
		Handler:      gw,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // EAS sync can be long-lived
		IdleTimeout:  90 * time.Second,
	}

	// TLS: use provided certificate if configured, otherwise plain HTTP
	// (plain HTTP is only acceptable behind an L4 TLS terminator)
	if cfg.Gateway.TLSCert != "" && cfg.Gateway.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.Gateway.TLSCert, cfg.Gateway.TLSKey)
		if err != nil {
			return fmt.Errorf("loading TLS certificate: %w", err)
		}
		srv.TLSConfig = &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		}
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("gateway listening",
			zap.String("addr", cfg.Gateway.Listen),
			zap.String("exchange", cfg.Gateway.ExchangeURL),
			zap.Bool("tls", srv.TLSConfig != nil),
		)

		var listenErr error
		if srv.TLSConfig != nil {
			listenErr = srv.ListenAndServeTLS("", "")
		} else {
			log.Warn("TLS not configured — gateway running in plain HTTP mode (only safe behind L4 TLS terminator)")
			listenErr = srv.ListenAndServe()
		}
		if !errors.Is(listenErr, http.ErrServerClosed) {
			errCh <- listenErr
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("gateway error: %w", err)
	case <-ctx.Done():
		log.Info("shutting down gateway")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("gateway shutdown: %w", err)
		}
	}

	return nil
}
