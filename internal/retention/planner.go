package retention

import (
	"fmt"
	"time"
)

type Action string

const (
	ActionCreate  Action = "create"
	ActionDetach  Action = "detach"
	ActionArchive Action = "archive"
	ActionDrop    Action = "drop"
)

type Config struct {
	CreatePartitionsAheadDays int
	RawRetentionDays          int
	RawExpiredAction          Action
	MinuteRetentionDays       int
	HourRetentionDays         int
	DayRetentionDays          int
}

type Partition struct {
	TableName     string
	PartitionDate time.Time
	Status        string
	ArchiveURI    string
	ErrorMessage  string
}

type PartitionAction struct {
	Action        Action
	TableName     string
	PartitionDate time.Time
}

type AggregateRetention struct {
	Before time.Time
}

type Planner struct {
	cfg Config
}

func NewPlanner(cfg Config) *Planner {
	if cfg.CreatePartitionsAheadDays == 0 {
		cfg.CreatePartitionsAheadDays = 7
	}
	if cfg.RawRetentionDays == 0 {
		cfg.RawRetentionDays = 30
	}
	if cfg.RawExpiredAction == "" {
		cfg.RawExpiredAction = ActionArchive
	}
	return &Planner{cfg: cfg}
}

func (p *Planner) PlanPartitionCreation(now time.Time) []PartitionAction {
	start := midnight(now)
	actions := make([]PartitionAction, 0, p.cfg.CreatePartitionsAheadDays)
	for i := 0; i < p.cfg.CreatePartitionsAheadDays; i++ {
		date := start.AddDate(0, 0, i)
		actions = append(actions, PartitionAction{
			Action:        ActionCreate,
			TableName:     fmt.Sprintf("raw_events_%s", date.Format("2006_01_02")),
			PartitionDate: date,
		})
	}
	return actions
}

func (p *Planner) PlanRawRetention(now time.Time, partitions []Partition) []PartitionAction {
	cutoff := midnight(now).AddDate(0, 0, -p.cfg.RawRetentionDays)
	var actions []PartitionAction
	for _, partition := range partitions {
		if partition.PartitionDate.Before(cutoff) {
			actions = append(actions, PartitionAction{
				Action:        p.cfg.RawExpiredAction,
				TableName:     partition.TableName,
				PartitionDate: partition.PartitionDate,
			})
		}
	}
	return actions
}

func (p *Planner) PlanAggregateRetention(now time.Time) map[string]AggregateRetention {
	result := make(map[string]AggregateRetention)
	if p.cfg.MinuteRetentionDays > 0 {
		result["minute"] = AggregateRetention{Before: midnight(now).AddDate(0, 0, -p.cfg.MinuteRetentionDays)}
	}
	if p.cfg.HourRetentionDays > 0 {
		result["hour"] = AggregateRetention{Before: midnight(now).AddDate(0, 0, -p.cfg.HourRetentionDays)}
	}
	if p.cfg.DayRetentionDays > 0 {
		result["day"] = AggregateRetention{Before: midnight(now).AddDate(0, 0, -p.cfg.DayRetentionDays)}
	}
	return result
}

func midnight(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
