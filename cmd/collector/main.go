package main

import (
	"log/slog"
	"net/http"
	"os"

	"webanalytics/internal/collector"
	"webanalytics/internal/config"
	"webanalytics/internal/platform/httpx"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/healthz", httpx.HealthHandler("collector"))
	credentials := collector.NewMemoryCredentials()
	collectorHandler := collector.NewHandler(collector.NewService(collector.ServiceDeps{
		Credentials: credentials,
		Publisher:   collector.NoopPublisher{},
		Online:      collector.NewMemoryOnlineTracker(),
		Limiter:     collector.AllowAllLimiter{},
	}))
	mux.HandleFunc("POST /collect", collectorHandler.Collect)

	slog.Info("collector listening", "addr", cfg.HTTP.CollectorAddr)
	if err := http.ListenAndServe(cfg.HTTP.CollectorAddr, httpx.RequestIDMiddleware(mux)); err != nil {
		slog.Error("collector stopped", "error", err)
		os.Exit(1)
	}
}
