package handler

import (
	"collab-consumer/internal/cache"
	"collab-consumer/internal/event"
	"context"
	"log"
)

type AssetChangesHandler struct {
	cache *cache.AssetCache
}

func NewAssetChangesHandler(cache *cache.AssetCache) *AssetChangesHandler {
	return &AssetChangesHandler{cache: cache}
}

func (h *AssetChangesHandler) HandleAssetEvent(ctx context.Context, e *event.Event) error {
	switch e.EventType {
	case "FOLDER_CREATED":
		if err := h.cache.SetUserAccess(ctx, e.AssetID, e.OwnerID, "OWNER"); err != nil {
			log.Printf("⚠️ Failed to set user access for FOLDER_CREATED: %v", err)
		}
	case "FOLDER_UPDATED":
		if err := h.cache.SetUserAccess(ctx, e.AssetID, e.TargetUser, e.AssetType); err != nil {
			log.Printf("⚠️ Failed to set user access for FOLDER_UPDATED: %v", err)
		}
	case "FOLDER_DELETED":
		if err := h.cache.RemoveUserAccess(ctx, e.AssetID, e.TargetUser); err != nil {
			log.Printf("⚠️ Failed to e user access for FOLDER_DELETED: %v", err)
		}
	case "FOLDER_SHARED":
		if err := h.cache.SetUserAccess(ctx, e.AssetID, e.TargetUser, e.AssetType); err != nil {
			log.Printf("⚠️ Failed to set user access for FOLDER_SHARED: %v", err)
		}
	case "FOLDER_UNSHARED":
		if err := h.cache.RemoveUserAccess(ctx, e.AssetID, e.TargetUser); err != nil {
			log.Printf("⚠️ Failed to remove user access for FOLDER_UNSHARED: %v", err)
		}
	case "NOTE_CREATED":
		if err := h.cache.SetUserAccess(ctx, e.AssetID, e.OwnerID, e.AssetType); err != nil {
			log.Printf("⚠️ Failed to set user access for NOTE_CREATED: %v", err)
		}
	case "NOTE_UPDATED":
		if err := h.cache.SetUserAccess(ctx, e.AssetID, e.TargetUser, e.AssetType); err != nil {
			log.Printf("⚠️ Failed to set user access for NOTE_UPDATED: %v", err)
		}
	case "NOTE_DELETED":
		if err := h.cache.RemoveUserAccess(ctx, e.AssetID, e.TargetUser); err != nil {
			log.Printf("⚠️ Failed to remove user access for NOTE_DELETED: %v", err)
		}
	case "NOTE_SHARED":
		if err := h.cache.SetUserAccess(ctx, e.AssetID, e.TargetUser, e.AssetType); err != nil {
			log.Printf("⚠️ Failed to set user access for NOTE_SHARED: %v", err)
		}
	case "NOTE_UNSHARED":
		if err := h.cache.RemoveUserAccess(ctx, e.AssetID, e.TargetUser); err != nil {
			log.Printf("⚠️ Failed to remove user access for NOTE_UNSHARED: %v", err)
		}
	default:
		log.Printf("⚠️ Unknown event: %s", e.EventType)
	}
	return nil
}
