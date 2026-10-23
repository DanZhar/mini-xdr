package main

import (
	"log/slog"
	"net/http"
	"os"
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
	cfg := loadConfig()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	logger.Info("starting server", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
