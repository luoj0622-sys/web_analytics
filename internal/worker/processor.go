package worker

import (
	"context"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
	"webanalytics/internal/store"
)

type Config struct {
	BatchSize   int
	IdleTimeout time.Duration
}

type ProcessorDeps struct {
	Consumer queue.EventConsumer
	Events   store.EventStore
	Stats    store.StatsStore
	Config   Config
}

type Processor struct {
	deps ProcessorDeps
}

func NewProcessor(deps ProcessorDeps) *Processor {
	if deps.Config.BatchSize == 0 {
		deps.Config.BatchSize = 1000
	}
	if deps.Config.IdleTimeout == 0 {
		deps.Config.IdleTimeout = time.Second
	}
	return &Processor{deps: deps}
}

func (p *Processor) RunOnce(ctx context.Context) error {
	batch, err := p.deps.Consumer.ConsumeBatch(ctx, queue.ConsumeOptions{
		MaxMessages: p.deps.Config.BatchSize,
		BlockFor:    p.deps.Config.IdleTimeout,
	})
	if err != nil {
		return err
	}
	if len(batch.Messages) == 0 {
		return nil
	}

	events := make([]domain.EventEnvelope, 0, len(batch.Messages))
	ids := make([]queue.MessageID, 0, len(batch.Messages))
	for _, message := range batch.Messages {
		events = append(events, message.Event)
		ids = append(ids, message.ID)
	}

	if err := p.deps.Events.WriteRawBatch(ctx, events); err != nil {
		_ = p.deps.Consumer.Retry(ctx, ids)
		return err
	}
	if err := p.deps.Stats.UpsertSiteStats(ctx, aggregateSiteStats(events)); err != nil {
		_ = p.deps.Consumer.Retry(ctx, ids)
		return err
	}
	if err := p.deps.Stats.UpsertDimensionStats(ctx, aggregateDimensionStats(events)); err != nil {
		_ = p.deps.Consumer.Retry(ctx, ids)
		return err
	}
	return p.deps.Consumer.Ack(ctx, ids)
}

func aggregateSiteStats(events []domain.EventEnvelope) []store.SiteStat {
	type key struct {
		siteID string
		bucket time.Time
	}
	seenVisitors := make(map[key]map[string]struct{})
	seenSessions := make(map[key]map[string]struct{})
	seenIPs := make(map[key]map[string]struct{})
	stats := make(map[key]store.SiteStat)

	for _, event := range events {
		bucket := event.OccurredAt.UTC().Truncate(time.Hour)
		k := key{siteID: event.SiteID, bucket: bucket}
		stat := stats[k]
		stat.SiteID = event.SiteID
		stat.Grain = store.GrainHour
		stat.Bucket = bucket

		if event.Type == domain.EventTypePageView {
			stat.PageViews++
		}
		if event.Type == domain.EventTypeCustom {
			stat.CustomEvents++
		}

		if seenVisitors[k] == nil {
			seenVisitors[k] = make(map[string]struct{})
		}
		if event.Visitor.ID != "" {
			seenVisitors[k][event.Visitor.ID] = struct{}{}
		}
		if seenSessions[k] == nil {
			seenSessions[k] = make(map[string]struct{})
		}
		if event.Visitor.SessionID != "" {
			seenSessions[k][event.Visitor.SessionID] = struct{}{}
		}
		if seenIPs[k] == nil {
			seenIPs[k] = make(map[string]struct{})
		}
		if event.Network.IP != "" {
			seenIPs[k][event.Network.IP] = struct{}{}
		}
		stats[k] = stat
	}

	out := make([]store.SiteStat, 0, len(stats))
	for k, stat := range stats {
		stat.UniqueVisitors = int64(len(seenVisitors[k]))
		stat.Sessions = int64(len(seenSessions[k]))
		stat.IPCount = int64(len(seenIPs[k]))
		out = append(out, stat)
	}
	return out
}

func aggregateDimensionStats(events []domain.EventEnvelope) []store.DimensionStat {
	type key struct {
		siteID    string
		dimension store.Dimension
		bucket    time.Time
		value     string
	}
	stats := make(map[key]store.DimensionStat)
	seenVisitors := make(map[key]map[string]struct{})
	seenSessions := make(map[key]map[string]struct{})
	seenIPs := make(map[key]map[string]struct{})

	add := func(event domain.EventEnvelope, dimension store.Dimension, value string, pageViews, eventCount int64) {
		if value == "" {
			return
		}
		bucket := time.Date(event.OccurredAt.UTC().Year(), event.OccurredAt.UTC().Month(), event.OccurredAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
		k := key{siteID: event.SiteID, dimension: dimension, bucket: bucket, value: value}
		stat := stats[k]
		stat.SiteID = event.SiteID
		stat.Dimension = dimension
		stat.Bucket = bucket
		stat.Key = value
		stat.PageViews += pageViews
		stat.EventCount += eventCount
		if seenVisitors[k] == nil {
			seenVisitors[k] = make(map[string]struct{})
		}
		if event.Visitor.ID != "" {
			seenVisitors[k][event.Visitor.ID] = struct{}{}
		}
		if seenSessions[k] == nil {
			seenSessions[k] = make(map[string]struct{})
		}
		if event.Visitor.SessionID != "" {
			seenSessions[k][event.Visitor.SessionID] = struct{}{}
		}
		if seenIPs[k] == nil {
			seenIPs[k] = make(map[string]struct{})
		}
		if event.Network.IP != "" {
			seenIPs[k][event.Network.IP] = struct{}{}
		}
		stats[k] = stat
	}

	for _, event := range events {
		if event.Type == domain.EventTypePageView {
			add(event, store.DimensionPage, event.Page.Path, 1, 0)
			add(event, store.DimensionReferrer, event.Page.Referrer, 1, 0)
			add(event, store.DimensionDevice, event.Device.Type, 1, 0)
			add(event, store.DimensionGeo, event.Network.Country, 1, 0)
		}
		if event.Type == domain.EventTypeCustom {
			add(event, store.DimensionEvent, event.Name, 0, 1)
		}
	}

	out := make([]store.DimensionStat, 0, len(stats))
	for k, stat := range stats {
		stat.UniqueVisitors = int64(len(seenVisitors[k]))
		stat.Sessions = int64(len(seenSessions[k]))
		stat.IPCount = int64(len(seenIPs[k]))
		out = append(out, stat)
	}
	return out
}
