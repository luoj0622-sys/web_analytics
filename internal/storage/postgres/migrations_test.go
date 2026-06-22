package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsDefineCoreBusinessTables(t *testing.T) {
	sql := readMigration(t, "001_core.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS sites",
		"CREATE TABLE IF NOT EXISTS tracking_credentials",
		"CREATE TABLE IF NOT EXISTS site_settings",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("001_core.sql missing %q", fragment)
		}
	}
}

func TestMigrationsDefinePartitionedRawEventsWithBoundedIndexes(t *testing.T) {
	sql := readMigration(t, "002_raw_events.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS raw_events",
		"PARTITION BY RANGE (event_date)",
		"CREATE INDEX IF NOT EXISTS idx_raw_events_site_time",
		"CREATE INDEX IF NOT EXISTS brin_raw_events_event_time",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("002_raw_events.sql missing %q", fragment)
		}
	}
}

func TestMigrationsDefineAggregateAndRetentionTables(t *testing.T) {
	sql := readMigration(t, "003_aggregates_retention.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS stats_site_minute",
		"CREATE TABLE IF NOT EXISTS stats_site_hour",
		"CREATE TABLE IF NOT EXISTS stats_site_day",
		"CREATE TABLE IF NOT EXISTS stats_page_day",
		"CREATE TABLE IF NOT EXISTS stats_referrer_day",
		"CREATE TABLE IF NOT EXISTS stats_device_day",
		"CREATE TABLE IF NOT EXISTS stats_event_day",
		"CREATE TABLE IF NOT EXISTS partition_lifecycle",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("003_aggregates_retention.sql missing %q", fragment)
		}
	}
}

func readMigration(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "migrations", name)
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(bytes)
}
