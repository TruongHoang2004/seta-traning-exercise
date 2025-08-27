package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID uuid.UUID

	Title string
	Body  string

	FolderID uuid.UUID
	Folder   *Folder

	Shared []*NoteShare

	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note, userID uuid.UUID) (*Note, error)
	GetOwner(ctx context.Context, noteID uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Note, error)
	GetByFolderID(ctx context.Context, folderID uuid.UUID) ([]*Note, error)
	GetAllCanAccess(ctx context.Context, userID uuid.UUID) ([]*Note, error)
	GetFolderAccessLevel(ctx context.Context, folderID, userID uuid.UUID) (AccessLevel, error)
	GetAccessLevel(ctx context.Context, noteID, userID uuid.UUID) (AccessLevel, error)
	ShareNote(ctx context.Context, noteID, userID uuid.UUID, accessLevel AccessLevel) error
	RevokeAccess(ctx context.Context, noteID, userID uuid.UUID) error
	ChangeAccessLevel(ctx context.Context, userID, folderID uuid.UUID, accessLevel AccessLevel) error
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id uuid.UUID) error
}
