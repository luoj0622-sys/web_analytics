package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/store"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) error
}

type EventStore struct {
	exec Executor
}

func NewEventStore(exec Executor) *EventStore {
	return &EventStore{exec: exec}
}

func (s *EventStore) WriteRawBatch(ctx context.Context, events []domain.EventEnvelope) error {
	if len(events) == 0 {
		return nil
	}

	columns := []string{
		"id", "site_id", "event_type", "event_name", "event_time", "event_date", "received_at",
		"visitor_id", "session_id", "page_url", "page_path", "page_title", "referrer",
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"browser", "os", "device_type", "country", "region", "city", "ip_hash", "user_agent",
	}

	args := make([]any, 0, len(events)*len(columns))
	rows := make([]string, 0, len(events))
	for i, event := range events {
		receivedAt := event.ReceivedAt
		if receivedAt.IsZero() {
			receivedAt = time.Now().UTC()
		}
		eventDate := event.OccurredAt.UTC().Format("2006-01-02")
		offset := i*len(columns) + 1
		rows = append(rows, placeholders(offset, len(columns)))
		args = append(args,
			event.ID,
			event.SiteID,
			string(event.Type),
			event.Name,
			event.OccurredAt.UTC(),
			eventDate,
			receivedAt.UTC(),
			event.Visitor.ID,
			event.Visitor.SessionID,
			event.Page.URL,
			event.Page.Path,
			event.Page.Title,
			event.Page.Referrer,
			event.Campaign.Source,
			event.Campaign.Medium,
			event.Campaign.Campaign,
			event.Campaign.Term,
			event.Campaign.Content,
			event.Device.Browser,
			event.Device.OS,
			event.Device.Type,
			event.Network.Country,
			event.Network.Region,
			event.Network.City,
			event.Network.IP,
			event.Network.UserAgent,
		)
	}

	query := "INSERT INTO raw_events (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(rows, ", ")
	return s.exec.ExecContext(ctx, query, args...)
}

type StatsStore struct {
	exec Executor
}

func NewStatsStore(exec Executor) *StatsStore {
	return &StatsStore{exec: exec}
}

func (s *StatsStore) UpsertSiteStats(ctx context.Context, stats []store.SiteStat) error {
	if len(stats) == 0 {
		return nil
	}

	grouped := map[store.Grain][]store.SiteStat{}
	for _, stat := range stats {
		grouped[stat.Grain] = append(grouped[stat.Grain], stat)
	}

	for grain, rowsForGrain := range grouped {
		if err := s.upsertSiteStatsForGrain(ctx, grain, rowsForGrain); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsStore) UpsertDimensionStats(ctx context.Context, stats []store.DimensionStat) error {
	if len(stats) == 0 {
		return nil
	}
	for _, stat := range stats {
		table, column, err := dimensionTarget(stat.Dimension)
		if err != nil {
			return err
		}
		query := fmt.Sprintf(`INSERT INTO %s (site_id, bucket, %s, page_views, unique_visitors, event_count)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING`, table, column)
		if err := s.exec.ExecContext(ctx, query, stat.SiteID, stat.Bucket.UTC(), stat.Key, stat.PageViews, stat.UniqueVisitors, stat.EventCount); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsStore) QuerySiteStats(context.Context, store.SiteStatsQuery) (store.SiteStatsReport, error) {
	return store.SiteStatsReport{}, fmt.Errorf("postgres stats query not implemented")
}

func (s *StatsStore) QueryDimensionStats(context.Context, store.DimensionStatsQuery) (store.DimensionStatsReport, error) {
	return store.DimensionStatsReport{}, fmt.Errorf("postgres dimension query not implemented")
}

func (s *StatsStore) upsertSiteStatsForGrain(ctx context.Context, grain store.Grain, stats []store.SiteStat) error {
	table, err := siteStatsTable(grain)
	if err != nil {
		return err
	}

	columns := []string{"site_id", "bucket", "page_views", "unique_visitors", "sessions", "custom_events"}
	args := make([]any, 0, len(stats)*len(columns))
	rows := make([]string, 0, len(stats))
	for i, stat := range stats {
		offset := i*len(columns) + 1
		rows = append(rows, placeholders(offset, len(columns)))
		args = append(args, stat.SiteID, stat.Bucket.UTC(), stat.PageViews, stat.UniqueVisitors, stat.Sessions, stat.CustomEvents)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s
ON CONFLICT (site_id, bucket) DO UPDATE SET
    page_views = %s.page_views + EXCLUDED.page_views,
    unique_visitors = %s.unique_visitors + EXCLUDED.unique_visitors,
    sessions = %s.sessions + EXCLUDED.sessions,
    custom_events = %s.custom_events + EXCLUDED.custom_events,
    updated_at = now()`,
		table,
		strings.Join(columns, ", "),
		strings.Join(rows, ", "),
		table,
		table,
		table,
		table,
	)
	return s.exec.ExecContext(ctx, query, args...)
}

func dimensionTarget(dimension store.Dimension) (string, string, error) {
	switch dimension {
	case store.DimensionPage:
		return "stats_page_day", "page_path", nil
	case store.DimensionReferrer:
		return "stats_referrer_day", "referrer", nil
	case store.DimensionDevice:
		return "stats_device_day", "device_type", nil
	case store.DimensionGeo:
		return "stats_geo_day", "country", nil
	case store.DimensionEvent:
		return "stats_event_day", "event_name", nil
	default:
		return "", "", fmt.Errorf("unsupported dimension %q", dimension)
	}
}

func siteStatsTable(grain store.Grain) (string, error) {
	switch grain {
	case store.GrainMinute:
		return "stats_site_minute", nil
	case store.GrainHour:
		return "stats_site_hour", nil
	case store.GrainDay:
		return "stats_site_day", nil
	default:
		return "", fmt.Errorf("unsupported grain %q", grain)
	}
}

func placeholders(offset, count int) string {
	parts := make([]string, count)
	for i := range count {
		parts[i] = fmt.Sprintf("$%d", offset+i)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
