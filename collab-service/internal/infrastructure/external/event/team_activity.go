package event

import (
	"collab-service/config"
	"collab-service/internal/infrastructure/external/event/kafka"
	"collab-service/internal/infrastructure/logger"
	"context"
	"encoding/json"
	"sync"
	"time"
)

type TeamEventType string

const (
	TeamCreated    TeamEventType = "TEAM_CREATED"
	MemberAdded    TeamEventType = "MEMBER_ADDED"
	MemberRemoved  TeamEventType = "MEMBER_REMOVED"
	ManagerAdded   TeamEventType = "MANAGER_ADDED"
	ManagerRemoved TeamEventType = "MANAGER_REMOVED"
)

type TeamEvent struct {
	EventType    TeamEventType `json:"eventType"`
	TeamID       string        `json:"teamId"`
	PerformedBy  string        `json:"performedBy"`
	TargetUserID string        `json:"targetUserId,omitempty"`
	Timestamp    string        `json:"timestamp"`
}

func NewTeamEvent(eventType TeamEventType, teamID, performedBy, targetUserID string) *TeamEvent {
	return &TeamEvent{
		EventType:    eventType,
		TeamID:       teamID,
		PerformedBy:  performedBy,
		TargetUserID: targetUserID,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
}

type TeamActivityProducer struct {
	Producer *kafka.Producer
}

var (
	teamActivityInstance *TeamActivityProducer
	teamActivityOnce     sync.Once
)

func GetTeamActivityProducer() *TeamActivityProducer {
	teamActivityOnce.Do(func() {
		logger.Info("Initializing TeamActivityProducer")
		teamActivityInstance = &TeamActivityProducer{
			Producer: kafka.NewProducer(config.GetConfig().KafkaAddresses, config.GetConfig().TeamActivityTopic),
		}
	})
	return teamActivityInstance
}

func (p *TeamActivityProducer) Produce(event *TeamEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	logger.Info("Producing team event", "event", event)
	return p.Producer.Publish(ctx, []byte(event.TeamID), eventJSON)
}
