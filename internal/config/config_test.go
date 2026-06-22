package config

import (
	"testing"
	"time"
)

func TestDefaultConfigDefinesAnalyticsRuntime(t *testing.T) {
	cfg := Default()

	if cfg.Queue.Driver != "redis-streams" {
		t.Fatalf("Queue.Driver = %q, want redis-streams", cfg.Queue.Driver)
	}
	if cfg.Storage.Driver != "postgres" {
		t.Fatalf("Storage.Driver = %q, want postgres", cfg.Storage.Driver)
	}
	if cfg.Worker.BatchSize != 1000 {
		t.Fatalf("Worker.BatchSize = %d, want 1000", cfg.Worker.BatchSize)
	}
	if cfg.Worker.FlushInterval != time.Second {
		t.Fatalf("Worker.FlushInterval = %s, want 1s", cfg.Worker.FlushInterval)
	}
	if cfg.Online.Window != 5*time.Minute {
		t.Fatalf("Online.Window = %s, want 5m", cfg.Online.Window)
	}
	if cfg.Retention.RawDays != 30 {
		t.Fatalf("Retention.RawDays = %d, want 30", cfg.Retention.RawDays)
	}
}

func TestLoadFromEnvOverridesDefaults(t *testing.T) {
	t.Setenv("WA_QUEUE_DRIVER", "rabbitmq")
	t.Setenv("WA_STORAGE_DRIVER", "clickhouse")
	t.Setenv("WA_WORKER_BATCH_SIZE", "2500")
	t.Setenv("WA_ONLINE_WINDOW", "10m")
	t.Setenv("WA_RAW_RETENTION_DAYS", "60")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.Queue.Driver != "rabbitmq" {
		t.Fatalf("Queue.Driver = %q, want rabbitmq", cfg.Queue.Driver)
	}
	if cfg.Storage.Driver != "clickhouse" {
		t.Fatalf("Storage.Driver = %q, want clickhouse", cfg.Storage.Driver)
	}
	if cfg.Worker.BatchSize != 2500 {
		t.Fatalf("Worker.BatchSize = %d, want 2500", cfg.Worker.BatchSize)
	}
	if cfg.Online.Window != 10*time.Minute {
		t.Fatalf("Online.Window = %s, want 10m", cfg.Online.Window)
	}
	if cfg.Retention.RawDays != 60 {
		t.Fatalf("Retention.RawDays = %d, want 60", cfg.Retention.RawDays)
	}
}
