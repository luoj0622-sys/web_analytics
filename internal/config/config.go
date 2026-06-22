package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP      HTTPConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Queue     QueueConfig
	Storage   StorageConfig
	Worker    WorkerConfig
	Online    OnlineConfig
	Retention RetentionConfig
}

type HTTPConfig struct {
	CollectorAddr string
	DashboardAddr string
}

type PostgresConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type QueueConfig struct {
	Driver        string
	Stream        string
	ConsumerGroup string
}

type StorageConfig struct {
	Driver string
}

type WorkerConfig struct {
	BatchSize     int
	FlushInterval time.Duration
}

type OnlineConfig struct {
	Window time.Duration
}

type RetentionConfig struct {
	RawDays          int
	MinuteAggDays    int
	HourAggDays      int
	DayAggDays       int
	CreatePartitions int
}

func Default() Config {
	return Config{
		HTTP: HTTPConfig{
			CollectorAddr: ":8080",
			DashboardAddr: ":8081",
		},
		Postgres: PostgresConfig{
			DSN: "postgres://webanalytics:webanalytics@localhost:5432/webanalytics?sslmode=disable",
		},
		Redis: RedisConfig{
			Addr: "localhost:6379",
		},
		Queue: QueueConfig{
			Driver:        "redis-streams",
			Stream:        "analytics:events",
			ConsumerGroup: "analytics-workers",
		},
		Storage: StorageConfig{
			Driver: "postgres",
		},
		Worker: WorkerConfig{
			BatchSize:     1000,
			FlushInterval: time.Second,
		},
		Online: OnlineConfig{
			Window: 5 * time.Minute,
		},
		Retention: RetentionConfig{
			RawDays:          30,
			MinuteAggDays:    15,
			HourAggDays:      365,
			DayAggDays:       0,
			CreatePartitions: 7,
		},
	}
}

func LoadFromEnv() (Config, error) {
	cfg := Default()

	cfg.HTTP.CollectorAddr = getenv("WA_COLLECTOR_ADDR", cfg.HTTP.CollectorAddr)
	cfg.HTTP.DashboardAddr = getenv("WA_DASHBOARD_ADDR", cfg.HTTP.DashboardAddr)
	cfg.Postgres.DSN = getenv("WA_POSTGRES_DSN", cfg.Postgres.DSN)
	cfg.Redis.Addr = getenv("WA_REDIS_ADDR", cfg.Redis.Addr)
	cfg.Redis.Password = getenv("WA_REDIS_PASSWORD", cfg.Redis.Password)
	cfg.Queue.Driver = getenv("WA_QUEUE_DRIVER", cfg.Queue.Driver)
	cfg.Queue.Stream = getenv("WA_QUEUE_STREAM", cfg.Queue.Stream)
	cfg.Queue.ConsumerGroup = getenv("WA_QUEUE_CONSUMER_GROUP", cfg.Queue.ConsumerGroup)
	cfg.Storage.Driver = getenv("WA_STORAGE_DRIVER", cfg.Storage.Driver)

	var err error
	if cfg.Redis.DB, err = getenvInt("WA_REDIS_DB", cfg.Redis.DB); err != nil {
		return Config{}, err
	}
	if cfg.Worker.BatchSize, err = getenvInt("WA_WORKER_BATCH_SIZE", cfg.Worker.BatchSize); err != nil {
		return Config{}, err
	}
	if cfg.Worker.FlushInterval, err = getenvDuration("WA_WORKER_FLUSH_INTERVAL", cfg.Worker.FlushInterval); err != nil {
		return Config{}, err
	}
	if cfg.Online.Window, err = getenvDuration("WA_ONLINE_WINDOW", cfg.Online.Window); err != nil {
		return Config{}, err
	}
	if cfg.Retention.RawDays, err = getenvInt("WA_RAW_RETENTION_DAYS", cfg.Retention.RawDays); err != nil {
		return Config{}, err
	}
	if cfg.Retention.MinuteAggDays, err = getenvInt("WA_MINUTE_AGG_RETENTION_DAYS", cfg.Retention.MinuteAggDays); err != nil {
		return Config{}, err
	}
	if cfg.Retention.HourAggDays, err = getenvInt("WA_HOUR_AGG_RETENTION_DAYS", cfg.Retention.HourAggDays); err != nil {
		return Config{}, err
	}
	if cfg.Retention.DayAggDays, err = getenvInt("WA_DAY_AGG_RETENTION_DAYS", cfg.Retention.DayAggDays); err != nil {
		return Config{}, err
	}
	if cfg.Retention.CreatePartitions, err = getenvInt("WA_CREATE_PARTITIONS_AHEAD_DAYS", cfg.Retention.CreatePartitions); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func getenvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
