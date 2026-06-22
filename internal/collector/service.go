package collector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"webanalytics/internal/domain"
	"webanalytics/internal/queue"
)

var ErrInvalidCredential = errors.New("invalid credential")

type Credential struct {
	SiteID  string
	Enabled bool
}

type CredentialValidator interface {
	Validate(context.Context, string, string) (Credential, error)
}

type OnlineTracker interface {
	MarkActive(ctx context.Context, siteID, visitorID string) error
}

type RateLimiter interface {
	Allow(ctx context.Context, siteID, visitorID string) (bool, error)
}

type AllowAllLimiter struct{}

func (AllowAllLimiter) Allow(context.Context, string, string) (bool, error) {
	return true, nil
}

type ServiceDeps struct {
	Credentials CredentialValidator
	Publisher   queue.EventPublisher
	Online      OnlineTracker
	Limiter     RateLimiter
	Now         func() time.Time
}

type Service struct {
	deps ServiceDeps
}

func NewService(deps ServiceDeps) *Service {
	if deps.Limiter == nil {
		deps.Limiter = AllowAllLimiter{}
	}
	if deps.Now == nil {
		deps.Now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{deps: deps}
}

type ClientMetadata struct {
	IP        string
	UserAgent string
}

type CollectRequest struct {
	PublicKey  string           `json:"public_key"`
	SiteID     string           `json:"site_id"`
	Type       domain.EventType `json:"type"`
	Name       string           `json:"name"`
	OccurredAt time.Time        `json:"occurred_at"`
	Visitor    domain.Visitor   `json:"visitor"`
	Page       domain.Page      `json:"page"`
	Campaign   domain.Campaign  `json:"campaign"`
	Device     domain.Device    `json:"device"`
	Properties map[string]any   `json:"properties"`
	Client     ClientMetadata   `json:"-"`
}

type CollectResult struct {
	Accepted  bool
	MessageID queue.MessageID
}

func (s *Service) Collect(ctx context.Context, req CollectRequest) (CollectResult, error) {
	credential, err := s.deps.Credentials.Validate(ctx, req.SiteID, req.PublicKey)
	if err != nil {
		return CollectResult{}, err
	}
	if !credential.Enabled {
		return CollectResult{}, ErrInvalidCredential
	}

	allowed, err := s.deps.Limiter.Allow(ctx, req.SiteID, req.Visitor.ID)
	if err != nil {
		return CollectResult{}, err
	}
	if !allowed {
		return CollectResult{}, fmt.Errorf("rate limited")
	}

	now := s.deps.Now()
	occurredAt := req.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}
	event := domain.EventEnvelope{
		ID:         "evt_" + randomSuffix(now),
		SiteID:     credential.SiteID,
		Type:       req.Type,
		Name:       req.Name,
		OccurredAt: occurredAt,
		ReceivedAt: now,
		Visitor:    req.Visitor,
		Page:       req.Page,
		Campaign:   req.Campaign,
		Device:     req.Device,
		Network: domain.Network{
			IP:        req.Client.IP,
			UserAgent: req.Client.UserAgent,
		},
		Properties: req.Properties,
	}

	msg, err := s.deps.Publisher.Publish(ctx, event)
	if err != nil {
		return CollectResult{}, err
	}
	if s.deps.Online != nil && req.Visitor.ID != "" && (req.Type == domain.EventTypePageView || req.Type == domain.EventTypeHeartbeat) {
		if err := s.deps.Online.MarkActive(ctx, credential.SiteID, req.Visitor.ID); err != nil {
			return CollectResult{}, err
		}
	}

	return CollectResult{Accepted: true, MessageID: msg.ID}, nil
}

func randomSuffix(now time.Time) string {
	return fmt.Sprintf("%d", now.UnixNano())
}
