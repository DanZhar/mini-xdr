package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func loadConfig() Config {
	config := Config{
		HTTPAddr:    ":8080",
		DatabaseURL: "postgres://xdr:xdr@localhost:5432/ti?sslmode=disable",
	}

	if v, ok := os.LookupEnv("HTTP_ADDR"); ok {
		config.HTTPAddr = v
	}

	if v, ok := os.LookupEnv("DATABASE_URL"); ok {
		config.DatabaseURL = v
	}

	return config
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	err := run(logger)

	if err != nil {
		logger.Error("ti-service failed", "err", err)
		os.Exit(1)
	}

}

func run(logger *slog.Logger) error {
	cfg := loadConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)

	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	errCh := make(chan error, 1)

	logger.Info("starting server", "addr", cfg.HTTPAddr)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
		logger.Info("shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		logger.Info("server stopped")
	}

	return nil
}

func healthz(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
