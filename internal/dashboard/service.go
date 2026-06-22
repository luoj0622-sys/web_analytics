package dashboard

import (
	"context"

	"webanalytics/internal/store"
)

type OnlineCounter interface {
	CountOnline(context.Context, string) (int64, error)
}

type StatsReader interface {
	QuerySiteStats(context.Context, store.SiteStatsQuery) (store.SiteStatsReport, error)
	QueryDimensionStats(context.Context, store.DimensionStatsQuery) (store.DimensionStatsReport, error)
}

type ServiceDeps struct {
	Online OnlineCounter
	Stats  StatsReader
}

type Service struct {
	deps ServiceDeps
}

func NewService(deps ServiceDeps) *Service {
	return &Service{deps: deps}
}

type Query struct {
	SiteID string
	Grain  store.Grain
}

type DimensionQuery struct {
	SiteID    string
	Dimension store.Dimension
}

type OnlineResult struct {
	SiteID string `json:"site_id"`
	Count  int64  `json:"count"`
}

type OverviewResult struct {
	SiteID         string `json:"site_id"`
	PageViews      int64  `json:"page_views"`
	UniqueVisitors int64  `json:"unique_visitors"`
	Sessions       int64  `json:"sessions"`
	CustomEvents   int64  `json:"custom_events"`
}

type TrendResult struct {
	SiteID string           `json:"site_id"`
	Grain  store.Grain      `json:"grain"`
	Rows   []store.SiteStat `json:"rows"`
}

type DimensionResult struct {
	SiteID    string                `json:"site_id"`
	Dimension store.Dimension       `json:"dimension"`
	Rows      []store.DimensionStat `json:"rows"`
}

func (s *Service) Online(ctx context.Context, siteID string) (OnlineResult, error) {
	count, err := s.deps.Online.CountOnline(ctx, siteID)
	if err != nil {
		return OnlineResult{}, err
	}
	return OnlineResult{SiteID: siteID, Count: count}, nil
}

func (s *Service) Overview(ctx context.Context, query Query) (OverviewResult, error) {
	report, err := s.deps.Stats.QuerySiteStats(ctx, store.SiteStatsQuery{SiteID: query.SiteID, Grain: query.Grain})
	if err != nil {
		return OverviewResult{}, err
	}
	result := OverviewResult{SiteID: query.SiteID}
	for _, row := range report.Rows {
		result.PageViews += row.PageViews
		result.UniqueVisitors += row.UniqueVisitors
		result.Sessions += row.Sessions
		result.CustomEvents += row.CustomEvents
	}
	return result, nil
}

func (s *Service) Trend(ctx context.Context, query Query) (TrendResult, error) {
	report, err := s.deps.Stats.QuerySiteStats(ctx, store.SiteStatsQuery{SiteID: query.SiteID, Grain: query.Grain})
	if err != nil {
		return TrendResult{}, err
	}
	return TrendResult{SiteID: report.SiteID, Grain: query.Grain, Rows: report.Rows}, nil
}

func (s *Service) DimensionReport(ctx context.Context, query DimensionQuery) (DimensionResult, error) {
	report, err := s.deps.Stats.QueryDimensionStats(ctx, store.DimensionStatsQuery{SiteID: query.SiteID, Dimension: query.Dimension})
	if err != nil {
		return DimensionResult{}, err
	}
	return DimensionResult{SiteID: report.SiteID, Dimension: report.Dimension, Rows: report.Rows}, nil
}
