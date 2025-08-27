package dto

import (
	"collab-service/internal/domain/entity"

	"github.com/google/uuid"
)

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}

type ShareFolderRequest struct {
	UserID      uuid.UUID          `json:"user_id" binding:"required"`
	AccessLevel entity.AccessLevel `json:"access_level" binding:"required"`
}

type FolderResponse struct {
	ID     uuid.UUID            `json:"id"`
	Name   string               `json:"name"`
	Shared []entity.FolderShare `json:"shared,omitempty"`
}
