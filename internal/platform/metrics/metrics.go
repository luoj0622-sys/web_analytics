package metrics

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type Tags map[string]string

type Recorder interface {
	IncCounter(name string, tags Tags)
	ObserveDuration(name string, value time.Duration, tags Tags)
}

type MemoryRecorder struct {
	mu        sync.Mutex
	counters  map[string]int64
	durations map[string][]time.Duration
}

func NewMemoryRecorder() *MemoryRecorder {
	return &MemoryRecorder{
		counters:  make(map[string]int64),
		durations: make(map[string][]time.Duration),
	}
}

func (r *MemoryRecorder) IncCounter(name string, tags Tags) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[key(name, tags)]++
}

func (r *MemoryRecorder) ObserveDuration(name string, value time.Duration, tags Tags) {
	r.mu.Lock()
	defer r.mu.Unlock()
	metricKey := key(name, tags)
	r.durations[metricKey] = append(r.durations[metricKey], value)
}

func (r *MemoryRecorder) CounterValue(name string, tags Tags) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counters[key(name, tags)]
}

func (r *MemoryRecorder) DurationCount(name string, tags Tags) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.durations[key(name, tags)])
}

func key(name string, tags Tags) string {
	if len(tags) == 0 {
		return name
	}
	parts := make([]string, 0, len(tags))
	for tag, value := range tags {
		parts = append(parts, tag+"="+value)
	}
	sort.Strings(parts)
	return name + "{" + strings.Join(parts, ",") + "}"
}
