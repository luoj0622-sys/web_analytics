package domain

import "time"

type EventType string

const (
	EventTypePageView  EventType = "page_view"
	EventTypeHeartbeat EventType = "heartbeat"
	EventTypeCustom    EventType = "custom"
)

type EventEnvelope struct {
	ID         string
	SiteID     string
	Type       EventType
	Name       string
	OccurredAt time.Time
	ReceivedAt time.Time
	Visitor    Visitor
	Page       Page
	Campaign   Campaign
	Device     Device
	Network    Network
	Properties map[string]any
}

type Visitor struct {
	ID        string
	SessionID string
	IsNew     bool
}

type Page struct {
	URL      string
	Path     string
	Title    string
	Referrer string
}

type Campaign struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string
}

type Device struct {
	Browser string
	OS      string
	Type    string
}

type Network struct {
	IP        string
	UserAgent string
	Country   string
	Region    string
	City      string
}
