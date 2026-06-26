package dashboard

import (
	"context"
	"testing"
	"time"

	"webanalytics/internal/store"
)

func TestServiceReturnsOnlineOverviewTrendAndDimensions(t *testing.T) {
	stats := &fakeStatsReader{}
	service := NewService(ServiceDeps{
		Online: fakeOnlineCounter{count: 42},
		Stats:  stats,
	})

	online, err := service.Online(context.Background(), "site_1")
	if err != nil {
		t.Fatalf("Online() error = %v", err)
	}
	if online.Count != 42 {
		t.Fatalf("online = %d, want 42", online.Count)
	}

	overview, err := service.Overview(context.Background(), Query{SiteID: "site_1", Grain: store.GrainHour})
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.PageViews != 10 {
		t.Fatalf("page views = %d, want 10", overview.PageViews)
	}
	if overview.IPCount != 6 {
		t.Fatalf("ip count = %d, want 6", overview.IPCount)
	}
	if overview.ActiveVisitors != 42 {
		t.Fatalf("active visitors = %d, want 42", overview.ActiveVisitors)
	}
	if overview.BlendedVisitors != 42 {
		t.Fatalf("blended visitors = %d, want 42", overview.BlendedVisitors)
	}

	trend, err := service.Trend(context.Background(), Query{SiteID: "site_1", Grain: store.GrainHour})
	if err != nil {
		t.Fatalf("Trend() error = %v", err)
	}
	if len(trend.Rows) != 1 {
		t.Fatalf("trend rows = %d, want 1", len(trend.Rows))
	}

	dimensions, err := service.DimensionReport(context.Background(), DimensionQuery{
		SiteID:    "site_1",
		Dimension: store.DimensionPage,
	})
	if err != nil {
		t.Fatalf("DimensionReport() error = %v", err)
	}
	if dimensions.Rows[0].Key != "/pricing" {
		t.Fatalf("dimension key = %q, want /pricing", dimensions.Rows[0].Key)
	}
	if stats.rawScans != 0 {
		t.Fatal("dashboard service scanned raw events")
	}
}

type fakeOnlineCounter struct {
	count int64
}

func (f fakeOnlineCounter) CountOnline(context.Context, string) (int64, error) {
	return f.count, nil
}

type fakeStatsReader struct {
	rawScans int
}

func (f *fakeStatsReader) QuerySiteStats(context.Context, store.SiteStatsQuery) (store.SiteStatsReport, error) {
	return store.SiteStatsReport{
		SiteID: "site_1",
		Rows: []store.SiteStat{{
			SiteID:         "site_1",
			Grain:          store.GrainHour,
			Bucket:         time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC),
			PageViews:      10,
			IPCount:        6,
			UniqueVisitors: 7,
			Sessions:       8,
			CustomEvents:   2,
		}},
	}, nil
}

func (f *fakeStatsReader) QueryDimensionStats(context.Context, store.DimensionStatsQuery) (store.DimensionStatsReport, error) {
	return store.DimensionStatsReport{
		SiteID: "site_1",
		Rows: []store.DimensionStat{{
			SiteID:    "site_1",
			Dimension: store.DimensionPage,
			Key:       "/pricing",
			PageViews: 10,
			IPCount:   6,
		}},
	}, nil
}
