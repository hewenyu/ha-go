package hago

import (
	"time"
)

// State represents the state of a Home Assistant entity
type State struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
	Context     Context                `json:"context"`
}

// Context provides information about when a state was changed
type Context struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// Service represents a Home Assistant service
type Service struct {
	Domain  string                 `json:"domain"`
	Service string                 `json:"service"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Domain represents a Home Assistant domain with its services
type Domain struct {
	Domain   string              `json:"domain"`
	Services map[string][]string `json:"services"`
}

// Entity represents a Home Assistant entity with its metadata
type Entity struct {
	EntityID     string                 `json:"entity_id"`
	Name         string                 `json:"name"`
	State        string                 `json:"state"`
	Attributes   map[string]interface{} `json:"attributes"`
	LastChanged  time.Time              `json:"last_changed"`
	LastUpdated  time.Time              `json:"last_updated"`
	Context      Context                `json:"context"`
	DeviceID     string                 `json:"device_id,omitempty"`
	AreaID       string                 `json:"area_id,omitempty"`
	Platform     string                 `json:"platform,omitempty"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
}

// Event represents a Home Assistant event
type Event struct {
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
	Origin    string                 `json:"origin"`
	TimeFired time.Time              `json:"time_fired"`
	Context   Context                `json:"context"`
}

// Device represents a Home Assistant device
type Device struct {
	ID               string    `json:"id"`
	AreaID           string    `json:"area_id,omitempty"`
	Name             string    `json:"name"`
	Manufacturer     string    `json:"manufacturer,omitempty"`
	Model            string    `json:"model,omitempty"`
	Identifiers      []string  `json:"identifiers,omitempty"`
	ViaDevice        string    `json:"via_device,omitempty"`
	LastSeen         time.Time `json:"last_seen,omitempty"`
	NameByUser       string    `json:"name_by_user,omitempty"`
	SWVersion        string    `json:"sw_version,omitempty"`
	HWVersion        string    `json:"hw_version,omitempty"`
	EntryType        string    `json:"entry_type,omitempty"`
	DisabledBy       string    `json:"disabled_by,omitempty"`
	ConfigurationURL string    `json:"configuration_url,omitempty"`
}

// Area represents a Home Assistant area
type Area struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Picture string   `json:"picture,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

// User represents a Home Assistant user
type User struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	IsOwner     bool     `json:"is_owner"`
	IsAdmin     bool     `json:"is_admin"`
	Credentials []string `json:"credentials"`
}
