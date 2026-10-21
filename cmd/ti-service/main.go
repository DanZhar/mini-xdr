package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	logger.Info("starting server", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
