package main

import (
	"log/slog"
	"os"

	"webanalytics/internal/config"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	slog.Info("worker configured",
		"queue_driver", cfg.Queue.Driver,
		"storage_driver", cfg.Storage.Driver,
		"batch_size", cfg.Worker.BatchSize,
	)
}
