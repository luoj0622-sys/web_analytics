package dashboard

import (
	"context"
	"time"

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
	From   time.Time
	To     time.Time
}

type DimensionQuery struct {
	SiteID    string
	Dimension store.Dimension
	From      time.Time
	To        time.Time
	Limit     int
}

type OnlineResult struct {
	SiteID string `json:"site_id"`
	Count  int64  `json:"count"`
}

type OverviewResult struct {
	SiteID               string `json:"site_id"`
	PageViews            int64  `json:"page_views"`
	IPCount              int64  `json:"ip_count"`
	UniqueVisitors       int64  `json:"unique_visitors"`
	Sessions             int64  `json:"sessions"`
	CustomEvents         int64  `json:"custom_events"`
	ActiveVisitors       int64  `json:"active_visitors"`
	SummedActiveVisitors int64  `json:"summed_active_visitors"`
	CumulativeVisitors   int64  `json:"cumulative_visitors"`
	BlendedVisitors      int64  `json:"blended_visitors"`
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
	report, err := s.deps.Stats.QuerySiteStats(ctx, store.SiteStatsQuery{SiteID: query.SiteID, Grain: query.Grain, From: query.From, To: query.To})
	if err != nil {
		return OverviewResult{}, err
	}
	result := OverviewResult{SiteID: query.SiteID}
	for _, row := range report.Rows {
		result.PageViews += row.PageViews
		result.IPCount += row.IPCount
		result.UniqueVisitors += row.UniqueVisitors
		result.Sessions += row.Sessions
		result.CustomEvents += row.CustomEvents
	}
	if s.deps.Online != nil {
		result.ActiveVisitors, err = s.deps.Online.CountOnline(ctx, query.SiteID)
		if err != nil {
			return OverviewResult{}, err
		}
	}
	result.SummedActiveVisitors = result.UniqueVisitors
	result.CumulativeVisitors = result.UniqueVisitors
	result.BlendedVisitors = maxInt64(result.CumulativeVisitors, result.ActiveVisitors)
	return result, nil
}

func (s *Service) Trend(ctx context.Context, query Query) (TrendResult, error) {
	report, err := s.deps.Stats.QuerySiteStats(ctx, store.SiteStatsQuery{SiteID: query.SiteID, Grain: query.Grain, From: query.From, To: query.To})
	if err != nil {
		return TrendResult{}, err
	}
	return TrendResult{SiteID: report.SiteID, Grain: query.Grain, Rows: report.Rows}, nil
}

func (s *Service) DimensionReport(ctx context.Context, query DimensionQuery) (DimensionResult, error) {
	report, err := s.deps.Stats.QueryDimensionStats(ctx, store.DimensionStatsQuery{SiteID: query.SiteID, Dimension: query.Dimension, From: query.From, To: query.To, Limit: query.Limit})
	if err != nil {
		return DimensionResult{}, err
	}
	return DimensionResult{SiteID: report.SiteID, Dimension: report.Dimension, Rows: report.Rows}, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
