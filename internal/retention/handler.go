package retention

import (
	"context"
	"encoding/json"
	"net/http"
)

type PartitionInspector interface {
	ListPartitions(context.Context) ([]Partition, error)
}

type Handler struct {
	inspector PartitionInspector
}

func NewHandler(inspector PartitionInspector) *Handler {
	return &Handler{inspector: inspector}
}

func (h *Handler) Partitions(w http.ResponseWriter, r *http.Request) {
	partitions, err := h.inspector.ListPartitions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"partitions": partitions})
}
