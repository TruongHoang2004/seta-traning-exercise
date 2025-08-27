package event

import (
	"collab-service/config"
	"collab-service/internal/infrastructure/external/event/kafka"
	"context"
	"encoding/json"
	"sync"
	"time"
)

// EventType represents the type of asset event
type EventType string

const (
	// Folder events
	FolderCreated  EventType = "FOLDER_CREATED"
	FolderUpdated  EventType = "FOLDER_UPDATED"
	FolderDeleted  EventType = "FOLDER_DELETED"
	FolderShared   EventType = "FOLDER_SHARED"
	FolderUnshared EventType = "FOLDER_UNSHARED"

	// Note events
	NoteCreated  EventType = "NOTE_CREATED"
	NoteUpdated  EventType = "NOTE_UPDATED"
	NoteDeleted  EventType = "NOTE_DELETED"
	NoteShared   EventType = "NOTE_SHARED"
	NoteUnshared EventType = "NOTE_UNSHARED"
)

type AssetType string

const (
	Folder AssetType = "FOLDER"
	Note   AssetType = "NOTE"
)

type AssetEvent struct {
	EventType EventType `json:"eventType"`
	AssetType AssetType `json:"assetType"`
	AssetId   string    `json:"assetId"`
	OwnerId   string    `json:"ownerId"`
	ActionBy  string    `json:"actionBy"`
	Timestamp string    `json:"timestamp"`
}

func NewAssetEvent(eventType EventType, assetType AssetType, assetId, ownerId, actionBy, timestamp string) *AssetEvent {
	return &AssetEvent{
		EventType: eventType,
		AssetType: assetType,
		AssetId:   assetId,
		OwnerId:   ownerId,
		ActionBy:  actionBy,
		Timestamp: timestamp,
	}
}

type AssetChangeProducer struct {
	Producer *kafka.Producer
}

var (
	assetChangeInstance *AssetChangeProducer
	assetChangeOnce     sync.Once
)

func GetAssetChangeProducer() *AssetChangeProducer {
	assetChangeOnce.Do(func() {
		assetChangeInstance = &AssetChangeProducer{
			Producer: kafka.NewProducer(config.GetConfig().KafkaAddresses, config.GetConfig().AssetChangeTopic),
		}
	})
	return assetChangeInstance
}

func (p *AssetChangeProducer) Produce(event *AssetEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.Producer.Publish(ctx, []byte(event.AssetId), eventJSON)
}
