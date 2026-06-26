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

type QueryExecutor interface {
	Executor
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
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
		query := fmt.Sprintf(`INSERT INTO %s (site_id, bucket, %s, page_views, ip_count, unique_visitors, sessions, event_count)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT DO NOTHING`, table, column)
		if err := s.exec.ExecContext(ctx, query, stat.SiteID, stat.Bucket.UTC(), stat.Key, stat.PageViews, stat.IPCount, stat.UniqueVisitors, stat.Sessions, stat.EventCount); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsStore) QuerySiteStats(ctx context.Context, query store.SiteStatsQuery) (store.SiteStatsReport, error) {
	db, ok := s.exec.(QueryExecutor)
	if !ok {
		return store.SiteStatsReport{}, fmt.Errorf("postgres stats query requires query executor")
	}
	table, err := siteStatsTable(query.Grain)
	if err != nil {
		return store.SiteStatsReport{}, err
	}
	clauses := []string{"site_id = $1"}
	args := []any{query.SiteID}
	if !query.From.IsZero() {
		args = append(args, query.From.UTC())
		clauses = append(clauses, fmt.Sprintf("bucket >= $%d", len(args)))
	}
	if !query.To.IsZero() {
		args = append(args, query.To.UTC())
		clauses = append(clauses, fmt.Sprintf("bucket < $%d", len(args)))
	}
	sql := fmt.Sprintf(`SELECT site_id, bucket, page_views, ip_count, unique_visitors, sessions, custom_events
FROM %s
WHERE %s
ORDER BY bucket ASC`, table, strings.Join(clauses, " AND "))
	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return store.SiteStatsReport{}, err
	}
	defer rows.Close()

	report := store.SiteStatsReport{SiteID: query.SiteID, Grain: query.Grain}
	for rows.Next() {
		var row store.SiteStat
		if err := rows.Scan(&row.SiteID, &row.Bucket, &row.PageViews, &row.IPCount, &row.UniqueVisitors, &row.Sessions, &row.CustomEvents); err != nil {
			return store.SiteStatsReport{}, err
		}
		row.Grain = query.Grain
		report.Rows = append(report.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return store.SiteStatsReport{}, err
	}
	return report, nil
}

func (s *StatsStore) QueryDimensionStats(ctx context.Context, query store.DimensionStatsQuery) (store.DimensionStatsReport, error) {
	db, ok := s.exec.(QueryExecutor)
	if !ok {
		return store.DimensionStatsReport{}, fmt.Errorf("postgres dimension query requires query executor")
	}
	table, column, err := dimensionTarget(query.Dimension)
	if err != nil {
		return store.DimensionStatsReport{}, err
	}
	clauses := []string{"site_id = $1"}
	args := []any{query.SiteID}
	if !query.From.IsZero() {
		args = append(args, query.From.UTC())
		clauses = append(clauses, fmt.Sprintf("bucket >= $%d", len(args)))
	}
	if !query.To.IsZero() {
		args = append(args, query.To.UTC())
		clauses = append(clauses, fmt.Sprintf("bucket < $%d", len(args)))
	}
	limit := query.Limit
	if limit == 0 {
		limit = 50
	}
	args = append(args, limit)
	sql := fmt.Sprintf(`SELECT site_id, bucket, %s, page_views, ip_count, unique_visitors, sessions, event_count
FROM %s
WHERE %s
ORDER BY page_views DESC, unique_visitors DESC
LIMIT $%d`, column, table, strings.Join(clauses, " AND "), len(args))
	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return store.DimensionStatsReport{}, err
	}
	defer rows.Close()

	report := store.DimensionStatsReport{SiteID: query.SiteID, Dimension: query.Dimension}
	for rows.Next() {
		var row store.DimensionStat
		if err := rows.Scan(&row.SiteID, &row.Bucket, &row.Key, &row.PageViews, &row.IPCount, &row.UniqueVisitors, &row.Sessions, &row.EventCount); err != nil {
			return store.DimensionStatsReport{}, err
		}
		row.Dimension = query.Dimension
		report.Rows = append(report.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return store.DimensionStatsReport{}, err
	}
	return report, nil
}

func (s *StatsStore) upsertSiteStatsForGrain(ctx context.Context, grain store.Grain, stats []store.SiteStat) error {
	table, err := siteStatsTable(grain)
	if err != nil {
		return err
	}

	columns := []string{"site_id", "bucket", "page_views", "ip_count", "unique_visitors", "sessions", "custom_events"}
	args := make([]any, 0, len(stats)*len(columns))
	rows := make([]string, 0, len(stats))
	for i, stat := range stats {
		offset := i*len(columns) + 1
		rows = append(rows, placeholders(offset, len(columns)))
		args = append(args, stat.SiteID, stat.Bucket.UTC(), stat.PageViews, stat.IPCount, stat.UniqueVisitors, stat.Sessions, stat.CustomEvents)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s
ON CONFLICT (site_id, bucket) DO UPDATE SET
    page_views = %s.page_views + EXCLUDED.page_views,
    ip_count = %s.ip_count + EXCLUDED.ip_count,
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
