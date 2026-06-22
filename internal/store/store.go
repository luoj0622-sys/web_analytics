package store

import (
	"context"
	"time"

	"webanalytics/internal/domain"
)

type Grain string

const (
	GrainMinute Grain = "minute"
	GrainHour   Grain = "hour"
	GrainDay    Grain = "day"
)

type EventStore interface {
	WriteRawBatch(context.Context, []domain.EventEnvelope) error
}

type StatsStore interface {
	UpsertSiteStats(context.Context, []SiteStat) error
	UpsertDimensionStats(context.Context, []DimensionStat) error
	QuerySiteStats(context.Context, SiteStatsQuery) (SiteStatsReport, error)
	QueryDimensionStats(context.Context, DimensionStatsQuery) (DimensionStatsReport, error)
}

type SiteStat struct {
	SiteID         string
	Grain          Grain
	Bucket         time.Time
	PageViews      int64
	UniqueVisitors int64
	Sessions       int64
	CustomEvents   int64
}

type SiteStatsQuery struct {
	SiteID string
	Grain  Grain
	From   time.Time
	To     time.Time
}

type SiteStatsReport struct {
	SiteID string
	Grain  Grain
	Rows   []SiteStat
}

type Dimension string

const (
	DimensionPage     Dimension = "page"
	DimensionReferrer Dimension = "referrer"
	DimensionUTM      Dimension = "utm"
	DimensionDevice   Dimension = "device"
	DimensionGeo      Dimension = "geo"
	DimensionEvent    Dimension = "event"
)

type DimensionStat struct {
	SiteID         string
	Dimension      Dimension
	Bucket         time.Time
	Key            string
	PageViews      int64
	UniqueVisitors int64
	EventCount     int64
}

type DimensionStatsQuery struct {
	SiteID    string
	Dimension Dimension
	From      time.Time
	To        time.Time
	Limit     int
}

type DimensionStatsReport struct {
	SiteID    string
	Dimension Dimension
	Rows      []DimensionStat
}
