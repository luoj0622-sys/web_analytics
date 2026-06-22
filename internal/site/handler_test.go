package site

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateSiteHandlerReturnsTrackingSnippet(t *testing.T) {
	service := NewService(&fakeSiteStore{}, ServiceConfig{
		SDKURL:     "https://analytics.example.com/sdk.js",
		CollectURL: "https://analytics.example.com/collect",
	})
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/sites", bytes.NewBufferString(`{"name":"Example","domain":"example.com"}`))
	req.Header.Set("X-Owner-User-ID", "user_1")
	rec := httptest.NewRecorder()

	handler.CreateSite(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "snippet") {
		t.Fatalf("response missing snippet: %s", rec.Body.String())
	}
}

func TestCreateSiteHandlerRequiresOwner(t *testing.T) {
	service := NewService(&fakeSiteStore{}, ServiceConfig{})
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/sites", bytes.NewBufferString(`{"name":"Example","domain":"example.com"}`))
	rec := httptest.NewRecorder()

	handler.CreateSite(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

var _ SiteStore = (*fakeSiteStore)(nil)

func (f *fakeSiteStore) _keepContextUsed(context.Context) {}
