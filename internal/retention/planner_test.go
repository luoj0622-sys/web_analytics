package retention

import (
	"testing"
	"time"
)

func TestPlannerCreatesFutureDailyPartitions(t *testing.T) {
	planner := NewPlanner(Config{CreatePartitionsAheadDays: 3})
	now := time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC)

	actions := planner.PlanPartitionCreation(now)

	if len(actions) != 3 {
		t.Fatalf("actions = %d, want 3", len(actions))
	}
	if actions[0].TableName != "raw_events_2026_06_22" {
		t.Fatalf("first table = %q", actions[0].TableName)
	}
	if actions[2].PartitionDate.Format("2006-01-02") != "2026-06-24" {
		t.Fatalf("third date = %s", actions[2].PartitionDate)
	}
}

func TestPlannerSelectsExpiredRawPartitions(t *testing.T) {
	planner := NewPlanner(Config{RawRetentionDays: 30, RawExpiredAction: ActionArchive})
	now := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)

	actions := planner.PlanRawRetention(now, []Partition{
		{TableName: "raw_events_2026_05_01", PartitionDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
		{TableName: "raw_events_2026_06_01", PartitionDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	})

	if len(actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(actions))
	}
	if actions[0].Action != ActionArchive {
		t.Fatalf("action = %q, want archive", actions[0].Action)
	}
}

func TestPlannerKeepsLongTermDayAggregates(t *testing.T) {
	planner := NewPlanner(Config{MinuteRetentionDays: 15, HourRetentionDays: 365, DayRetentionDays: 0})
	now := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)

	actions := planner.PlanAggregateRetention(now)

	if actions["minute"].Before != now.AddDate(0, 0, -15) {
		t.Fatalf("minute cutoff = %s", actions["minute"].Before)
	}
	if _, ok := actions["day"]; ok {
		t.Fatal("day aggregates should be retained indefinitely")
	}
}
