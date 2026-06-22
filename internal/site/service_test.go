package site

import (
	"context"
	"strings"
	"testing"
)

func TestCreateSiteGeneratesTrackingSnippet(t *testing.T) {
	store := &fakeSiteStore{}
	service := NewService(store, ServiceConfig{
		SDKURL:     "https://analytics.example.com/sdk.js",
		CollectURL: "https://analytics.example.com/collect",
	})

	result, err := service.CreateSite(context.Background(), CreateSiteRequest{
		OwnerUserID: "user_1",
		Name:        "Example",
		Domain:      "example.com",
	})
	if err != nil {
		t.Fatalf("CreateSite() error = %v", err)
	}

	if result.Site.ID == "" {
		t.Fatal("site id is empty")
	}
	if result.Credential.PublicKey == "" {
		t.Fatal("public key is empty")
	}
	if !strings.Contains(result.Snippet, "https://analytics.example.com/sdk.js") {
		t.Fatalf("snippet missing sdk url: %s", result.Snippet)
	}
	if !strings.Contains(result.Snippet, result.Site.ID) {
		t.Fatalf("snippet missing site id: %s", result.Snippet)
	}
	if store.saved.Domain != "example.com" {
		t.Fatalf("saved domain = %q, want example.com", store.saved.Domain)
	}
}

type fakeSiteStore struct {
	saved Site
}

func (f *fakeSiteStore) CreateSite(_ context.Context, site Site, credential TrackingCredential) (Site, TrackingCredential, error) {
	f.saved = site
	return site, credential, nil
}
