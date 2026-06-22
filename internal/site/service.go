package site

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"time"
)

type Site struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"owner_user_id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type TrackingCredential struct {
	ID        string    `json:"id"`
	SiteID    string    `json:"site_id"`
	PublicKey string    `json:"public_key"`
	Secret    string    `json:"-"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateSiteRequest struct {
	OwnerUserID string `json:"-"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
}

type CreateSiteResult struct {
	Site       Site               `json:"site"`
	Credential TrackingCredential `json:"credential"`
	Snippet    string             `json:"snippet"`
}

type SiteStore interface {
	CreateSite(context.Context, Site, TrackingCredential) (Site, TrackingCredential, error)
}

type ServiceConfig struct {
	SDKURL     string
	CollectURL string
}

type Service struct {
	store SiteStore
	cfg   ServiceConfig
}

func NewService(store SiteStore, cfg ServiceConfig) *Service {
	if cfg.SDKURL == "" {
		cfg.SDKURL = "/sdk.js"
	}
	if cfg.CollectURL == "" {
		cfg.CollectURL = "/collect"
	}
	return &Service{store: store, cfg: cfg}
}

func (s *Service) CreateSite(ctx context.Context, req CreateSiteRequest) (CreateSiteResult, error) {
	if req.OwnerUserID == "" {
		return CreateSiteResult{}, fmt.Errorf("owner user id is required")
	}
	if req.Name == "" {
		return CreateSiteResult{}, fmt.Errorf("site name is required")
	}
	if req.Domain == "" {
		return CreateSiteResult{}, fmt.Errorf("site domain is required")
	}

	now := time.Now().UTC()
	site := Site{
		ID:          "site_" + randomHex(12),
		OwnerUserID: req.OwnerUserID,
		Name:        req.Name,
		Domain:      req.Domain,
		Enabled:     true,
		CreatedAt:   now,
	}
	credential := TrackingCredential{
		ID:        "cred_" + randomHex(12),
		SiteID:    site.ID,
		PublicKey: "pk_" + randomHex(16),
		Secret:    "sk_" + randomHex(32),
		Enabled:   true,
		CreatedAt: now,
	}

	savedSite, savedCredential, err := s.store.CreateSite(ctx, site, credential)
	if err != nil {
		return CreateSiteResult{}, err
	}

	return CreateSiteResult{
		Site:       savedSite,
		Credential: savedCredential,
		Snippet:    s.snippet(savedSite.ID, savedCredential.PublicKey),
	}, nil
}

func (s *Service) snippet(siteID, publicKey string) string {
	return fmt.Sprintf(`<script async src="%s" data-site-id="%s" data-public-key="%s" data-collect-url="%s"></script>`,
		html.EscapeString(s.cfg.SDKURL),
		html.EscapeString(siteID),
		html.EscapeString(publicKey),
		html.EscapeString(s.cfg.CollectURL),
	)
}

func randomHex(bytesLen int) string {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
