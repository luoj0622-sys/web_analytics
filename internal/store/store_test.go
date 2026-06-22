package store

import (
	"context"
	"testing"
	"time"

	"webanalytics/internal/domain"
)

func TestStoreInterfacesSeparateRawEventsAndAggregates(t *testing.T) {
	var events EventStore = noopEventStore{}
	var stats StatsStore = noopStatsStore{}

	ctx := context.Background()
	if err := events.WriteRawBatch(ctx, []domain.EventEnvelope{{ID: "evt_1"}}); err != nil {
		t.Fatalf("WriteRawBatch() error = %v", err)
	}
	if err := stats.UpsertSiteStats(ctx, []SiteStat{{SiteID: "site_1", Grain: GrainHour, Bucket: time.Now().UTC(), PageViews: 5}}); err != nil {
		t.Fatalf("UpsertSiteStats() error = %v", err)
	}
	if err := stats.UpsertDimensionStats(ctx, []DimensionStat{{SiteID: "site_1", Dimension: DimensionPage, Bucket: time.Now().UTC(), Key: "/pricing", PageViews: 5}}); err != nil {
		t.Fatalf("UpsertDimensionStats() error = %v", err)
	}
	report, err := stats.QuerySiteStats(ctx, SiteStatsQuery{SiteID: "site_1", Grain: GrainHour})
	if err != nil {
		t.Fatalf("QuerySiteStats() error = %v", err)
	}
	if report.SiteID != "site_1" {
		t.Fatalf("SiteID = %q, want site_1", report.SiteID)
	}
	dimReport, err := stats.QueryDimensionStats(ctx, DimensionStatsQuery{SiteID: "site_1", Dimension: DimensionPage})
	if err != nil {
		t.Fatalf("QueryDimensionStats() error = %v", err)
	}
	if dimReport.Dimension != DimensionPage {
		t.Fatalf("Dimension = %q, want page", dimReport.Dimension)
	}
}

type noopEventStore struct{}

func (noopEventStore) WriteRawBatch(context.Context, []domain.EventEnvelope) error {
	return nil
}

type noopStatsStore struct{}

func (noopStatsStore) UpsertSiteStats(context.Context, []SiteStat) error {
	return nil
}

func (noopStatsStore) UpsertDimensionStats(context.Context, []DimensionStat) error {
	return nil
}

func (noopStatsStore) QuerySiteStats(context.Context, SiteStatsQuery) (SiteStatsReport, error) {
	return SiteStatsReport{SiteID: "site_1"}, nil
}

func (noopStatsStore) QueryDimensionStats(context.Context, DimensionStatsQuery) (DimensionStatsReport, error) {
	return DimensionStatsReport{SiteID: "site_1", Dimension: DimensionPage}, nil
}
