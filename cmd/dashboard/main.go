package main

import (
	"log/slog"
	"net/http"
	"os"

	"webanalytics/internal/config"
	"webanalytics/internal/dashboard"
	"webanalytics/internal/platform/httpx"
	"webanalytics/internal/site"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/healthz", httpx.HealthHandler("dashboard"))
	siteHandler := site.NewHandler(site.NewService(site.NewMemoryStore(), site.ServiceConfig{}))
	mux.HandleFunc("POST /api/sites", siteHandler.CreateSite)
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(dashboard.ServiceDeps{
		Online: dashboard.MemoryOnlineCounter{},
		Stats:  dashboard.EmptyStatsReader{},
	}))
	mux.Handle("/api/sites/", dashboardHandler)

	slog.Info("dashboard listening", "addr", cfg.HTTP.DashboardAddr)
	if err := http.ListenAndServe(cfg.HTTP.DashboardAddr, httpx.RequestIDMiddleware(mux)); err != nil {
		slog.Error("dashboard stopped", "error", err)
		os.Exit(1)
	}
}
