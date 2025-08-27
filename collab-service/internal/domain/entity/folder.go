package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	ID   uuid.UUID
	Name string

	Notes  []Note
	Shared []FolderShare

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFolder(name string) *Folder {
	return &Folder{
		Name:      name,
		Notes:     []Note{},
		Shared:    []FolderShare{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type FolderRepository interface {
	Create(ctx context.Context, folder *Folder) (*Folder, error)
	GetOwner(ctx context.Context, folderID uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	GetAllForCanAccess(ctx context.Context, userID uuid.UUID) ([]*Folder, error)
	GetAccessLevel(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) (AccessLevel, error)
	ShareFolder(ctx context.Context, folderID, userID uuid.UUID, accessLevel AccessLevel) error
	RevokeAccess(ctx context.Context, folderID, userID uuid.UUID) error
	ChangeAccessLevel(ctx context.Context, folderID, userID uuid.UUID, accessLevel AccessLevel) error
	Update(ctx context.Context, folder *Folder) error
	Delete(ctx context.Context, id uuid.UUID) error
}
