package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"webanalytics/internal/store"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	siteID, action, ok := parseSiteAction(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var result any
	var err error
	switch action {
	case "online":
		result, err = h.service.Online(r.Context(), siteID)
	case "overview":
		from, to := timeRangeFromQuery(r)
		result, err = h.service.Overview(r.Context(), Query{SiteID: siteID, Grain: grainFromQuery(r), From: from, To: to})
	case "trend":
		from, to := timeRangeFromQuery(r)
		result, err = h.service.Trend(r.Context(), Query{SiteID: siteID, Grain: grainFromQuery(r), From: from, To: to})
	case "dimensions":
		from, to := timeRangeFromQuery(r)
		result, err = h.service.DimensionReport(r.Context(), DimensionQuery{SiteID: siteID, Dimension: dimensionFromQuery(r), From: from, To: to, Limit: limitFromQuery(r)})
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func parseSiteAction(path string) (siteID, action string, ok bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[0] != "api" || parts[1] != "sites" {
		return "", "", false
	}
	return parts[2], parts[3], true
}

func grainFromQuery(r *http.Request) store.Grain {
	switch r.URL.Query().Get("grain") {
	case "minute":
		return store.GrainMinute
	case "day":
		return store.GrainDay
	default:
		return store.GrainHour
	}
}

func timeRangeFromQuery(r *http.Request) (from, to time.Time) {
	if value := r.URL.Query().Get("from"); value != "" {
		from = parseQueryTime(value)
	}
	if value := r.URL.Query().Get("to"); value != "" {
		to = parseQueryTime(value)
	}
	return from, to
}

func parseQueryTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func dimensionFromQuery(r *http.Request) store.Dimension {
	switch r.URL.Query().Get("dimension") {
	case "referrer":
		return store.DimensionReferrer
	case "utm":
		return store.DimensionUTM
	case "device":
		return store.DimensionDevice
	case "geo":
		return store.DimensionGeo
	case "event":
		return store.DimensionEvent
	default:
		return store.DimensionPage
	}
}

func limitFromQuery(r *http.Request) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		return 0
	}
	return limit
}
