package ops

import (
	"os"
	"strings"
	"testing"
)

func TestOperationalAssetsDocumentCapacityAndBacklog(t *testing.T) {
	for _, path := range []string{
		"../../README.md",
		"../../config/local.example.yaml",
		"../../docs/observability.md",
		"../../docs/runbook.md",
		"../../tools/loadtest.mjs",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}

	observability := readFile(t, "../../docs/observability.md")
	for _, fragment := range []string{
		"collector_request_rate",
		"redis_stream_consumer_lag",
		"postgres_write_latency",
		"30 million events/day",
		"50 million events/day",
	} {
		if !strings.Contains(observability, fragment) {
			t.Fatalf("observability doc missing %q", fragment)
		}
	}
}

func TestReadmeAndConfigDocumentRuntimeDependencies(t *testing.T) {
	readme := readFile(t, "../../README.md")
	for _, fragment := range []string{
		"Redis Streams",
		"PostgreSQL",
		"config/local.example.yaml",
		"RabbitMQ",
		"ClickHouse",
	} {
		if !strings.Contains(readme, fragment) {
			t.Fatalf("README missing %q", fragment)
		}
	}

	config := readFile(t, "../../config/local.example.yaml")
	for _, fragment := range []string{
		"postgres:",
		"redis:",
		"queue:",
		"stream: analytics:events",
		"consumer_group: analytics-workers",
		"stable_events_per_day: 30000000",
		"stretch_events_per_day: 50000000",
	} {
		if !strings.Contains(config, fragment) {
			t.Fatalf("config missing %q", fragment)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(bytes)
}
