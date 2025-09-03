package handler

import (
	"collab-consumer/internal/cache"
	"collab-consumer/internal/event"
	"context"
	"log"
)

type TeamEventHandler struct {
	Cache *cache.TeamCache
}

func NewTeamEventHandler(cache *cache.TeamCache) *TeamEventHandler {
	return &TeamEventHandler{Cache: cache}
}

func (h *TeamEventHandler) HandleTeamEvent(ctx context.Context, e *event.Event) error {
	switch e.EventType {
	case "MEMBER_ADDED":
		if err := h.Cache.AddTeamMember(ctx, e.TeamID, e.TargetUser); err != nil {
			log.Printf("❌ Failed to add team member cache: %v", err)
		}
	case "MEMBER_REMOVED":
		if err := h.Cache.RemoveTeamMember(ctx, e.TeamID, e.TargetUser); err != nil {
			log.Printf("❌ Failed to remove team member cache: %v", err)
		}
	case "TEAM_CREATED":
		if err := h.Cache.AddTeamMember(ctx, e.TeamID, e.PerformedBy); err != nil {
			log.Printf("❌ Failed to add team creator to cache: %v", err)
		}
	case "MANAGER_ADDED":
		if err := h.Cache.AddTeamMember(ctx, e.TeamID, e.TargetUser); err != nil {
			log.Printf("❌ Failed to add team manager cache: %v", err)
		}
	case "MANAGER_REMOVED":
		if err := h.Cache.RemoveTeamMember(ctx, e.TeamID, e.TargetUser); err != nil {
			log.Printf("❌ Failed to remove team manager cache: %v", err)
		}
	default:
		log.Printf("⚠️ Unknown team event: %s", e.EventType)
	}
	return nil
}
