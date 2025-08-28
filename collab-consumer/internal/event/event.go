package event

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	EventType   string `json:"eventType"`
	TeamID      string `json:"teamId,omitempty"`
	PerformedBy string `json:"performedBy,omitempty"`
	TargetUser  string `json:"targetUserId,omitempty"`

	AssetType string `json:"assetType,omitempty"`
	AssetID   string `json:"assetId,omitempty"`
	OwnerID   string `json:"ownerId,omitempty"`
	ActionBy  string `json:"actionBy,omitempty"`

	Timestamp string `json:"timestamp"`
}

// Parse raw Kafka message
func ParseEvent(msg []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(msg, &e); err != nil {
		return nil, fmt.Errorf("invalid event: %w", err)
	}
	return &e, nil
}
