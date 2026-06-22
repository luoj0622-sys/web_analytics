package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"webanalytics/internal/store"
)

func TestHandlerServesOnlineOverviewTrendAndDimensions(t *testing.T) {
	service := NewService(ServiceDeps{
		Online: fakeOnlineCounter{count: 5},
		Stats:  &fakeStatsReader{},
	})
	handler := NewHandler(service)

	for _, tc := range []struct {
		path string
		want string
	}{
		{"/api/sites/site_1/online", `"count":5`},
		{"/api/sites/site_1/overview", `"page_views":10`},
		{"/api/sites/site_1/trend?grain=hour", `"rows"`},
		{"/api/sites/site_1/dimensions?dimension=page", `"/pricing"`},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200: %s", tc.path, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), tc.want) {
			t.Fatalf("%s body = %s, want fragment %s", tc.path, rec.Body.String(), tc.want)
		}
	}
}

var _ StatsReader = (*fakeStatsReader)(nil)
var _ OnlineCounter = fakeOnlineCounter{}
var _ = store.GrainHour
