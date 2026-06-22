package collector

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerCollectsPageView(t *testing.T) {
	service := NewService(ServiceDeps{
		Credentials: fakeCredentials{valid: true},
		Publisher:   &fakePublisher{},
		Online:      &fakeOnlineTracker{},
		Limiter:     AllowAllLimiter{},
	})
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/collect", bytes.NewBufferString(`{
		"site_id":"site_1",
		"public_key":"pk_1",
		"type":"page_view",
		"visitor":{"id":"visitor_1","session_id":"session_1"}
	}`))
	req.Header.Set("User-Agent", "Mozilla/5.0 Test")
	req.RemoteAddr = "203.0.113.1:1234"
	rec := httptest.NewRecorder()

	handler.Collect(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestHandlerRejectsInvalidCredential(t *testing.T) {
	service := NewService(ServiceDeps{
		Credentials: fakeCredentials{valid: false},
		Publisher:   &fakePublisher{},
		Online:      &fakeOnlineTracker{},
		Limiter:     AllowAllLimiter{},
	})
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/collect", bytes.NewBufferString(`{
		"site_id":"site_1",
		"public_key":"bad",
		"type":"page_view",
		"visitor":{"id":"visitor_1","session_id":"session_1"}
	}`))
	rec := httptest.NewRecorder()

	handler.Collect(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
