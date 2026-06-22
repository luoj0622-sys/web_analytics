package metrics

import (
	"testing"
	"time"
)

func TestMemoryRecorderTracksCountersAndDurations(t *testing.T) {
	recorder := NewMemoryRecorder()

	recorder.IncCounter("collector_requests_total", Tags{"service": "collector"})
	recorder.ObserveDuration("queue_publish_latency", 25*time.Millisecond, Tags{"driver": "redis-streams"})

	if got := recorder.CounterValue("collector_requests_total", Tags{"service": "collector"}); got != 1 {
		t.Fatalf("counter = %d, want 1", got)
	}
	if got := recorder.DurationCount("queue_publish_latency", Tags{"driver": "redis-streams"}); got != 1 {
		t.Fatalf("duration count = %d, want 1", got)
	}
}
