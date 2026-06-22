package retention

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandlerReportsPartitionStatus(t *testing.T) {
	handler := NewHandler(fakePartitionInspector{})
	req := httptest.NewRequest(http.MethodGet, "/admin/partitions", nil)
	rec := httptest.NewRecorder()

	handler.Partitions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "raw_events_2026_06_22") {
		t.Fatalf("body missing partition: %s", rec.Body.String())
	}
}

type fakePartitionInspector struct{}

func (fakePartitionInspector) ListPartitions(context.Context) ([]Partition, error) {
	return []Partition{{
		TableName:     "raw_events_2026_06_22",
		PartitionDate: time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC),
		Status:        "active",
	}}, nil
}
