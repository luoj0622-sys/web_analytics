package dashboard

import (
	"context"

	"webanalytics/internal/store"
)

type MemoryOnlineCounter struct{}

func (MemoryOnlineCounter) CountOnline(context.Context, string) (int64, error) {
	return 0, nil
}

type EmptyStatsReader struct{}

func (EmptyStatsReader) QuerySiteStats(_ context.Context, query store.SiteStatsQuery) (store.SiteStatsReport, error) {
	return store.SiteStatsReport{SiteID: query.SiteID, Grain: query.Grain}, nil
}

func (EmptyStatsReader) QueryDimensionStats(_ context.Context, query store.DimensionStatsQuery) (store.DimensionStatsReport, error) {
	return store.DimensionStatsReport{SiteID: query.SiteID, Dimension: query.Dimension}, nil
}
