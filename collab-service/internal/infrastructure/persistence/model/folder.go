package model

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type FolderModel struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

	Name string `gorm:"type:varchar(255);not null"`

	Notes  []NoteModel        `gorm:"foreignKey:FolderID;references:ID"`
	Shared []FolderShareModel `gorm:"foreignKey:FolderID;references:ID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (FolderModel) TableName() string {
	return "folders"
}

func (m *FolderModel) ToDomain() *entity.Folder {
	folder := &entity.Folder{
		ID:     m.ID,
		Name:   m.Name,
		Notes:  make([]entity.Note, len(m.Notes)),
		Shared: make([]entity.FolderShare, len(m.Shared)),
	}

	for i, note := range m.Notes {
		noteEntity := note.ToDomain()
		folder.Notes[i] = *noteEntity
	}

	for i, share := range m.Shared {
		shareEntity := share.ToDomain()
		folder.Shared[i] = *shareEntity
	}

	return folder
}

func FolderModelFromDomain(folderEntity *entity.Folder) *FolderModel {
	m := &FolderModel{
		ID:     folderEntity.ID,
		Name:   folderEntity.Name,
		Notes:  make([]NoteModel, len(folderEntity.Notes)),
		Shared: make([]FolderShareModel, len(folderEntity.Shared)),
	}
	for i, note := range folderEntity.Notes {
		m.Notes[i] = *NoteModelFromDomain(&note)
	}
	for i, share := range folderEntity.Shared {
		m.Shared[i] = *FolderShareModelFromDomain(&share)
	}
	return m
}
