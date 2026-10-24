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
)

type Config struct {
	HTTPAddr string
}

func loadConfig() Config {
	addr, ok := os.LookupEnv("HTTP_ADDR")

	config := Config{HTTPAddr: ":8080"}

	if ok {
		config.HTTPAddr = addr
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
