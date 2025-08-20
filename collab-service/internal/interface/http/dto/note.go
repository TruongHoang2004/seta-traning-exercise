package dto

import (
	"collab-service/internal/domain/entity"

	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	FolderID uuid.UUID `json:"folder_id"`
}

type UpdateNoteRequest struct {
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	FolderID uuid.UUID `json:"folder_id"`
}

type ShareNoteRequest struct {
	AccessLevel entity.AccessLevel `json:"access_level"`
	UserID      uuid.UUID          `json:"user_id"`
}

type NoteResponse struct {
	ID        uuid.UUID       `json:"id"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	FolderID  uuid.UUID       `json:"folder_id"`
	Folder    *FolderResponse `json:"folder,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}
