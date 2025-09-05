package model

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type NoteModel struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

	Title string `gorm:"type:varchar(255);not null"`
	Body  string `gorm:"type:text;not null"`

	FolderID uuid.UUID    `gorm:"type:uuid;not null;index"`
	Folder   *FolderModel `gorm:"foreignKey:FolderID;references:ID"`

	Shared []NoteShareModel `gorm:"foreignKey:NoteID;references:ID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (NoteModel) TableName() string {
	return "notes"
}

func (m *NoteModel) ToDomain() *entity.Note {
	var folder *entity.Folder
	if m.Folder != nil {
		folder = m.Folder.ToDomain()
	}

	shared := make([]*entity.NoteShare, len(m.Shared))
	for i, share := range m.Shared {
		shared[i] = share.ToDomain()
	}

	return &entity.Note{
		ID:        m.ID,
		Title:     m.Title,
		Body:      m.Body,
		FolderID:  m.FolderID,
		Folder:    folder,
		Shared:    shared,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func NoteModelFromDomain(noteEntity *entity.Note) *NoteModel {
	var folderModel *FolderModel
	if noteEntity.Folder != nil {
		folderModel = FolderModelFromDomain(noteEntity.Folder)
	}

	var shared []NoteShareModel
	for _, share := range noteEntity.Shared {
		shared = append(shared, *NoteShareModelFromDomain(share))
	}
	m := &NoteModel{
		ID:        noteEntity.ID,
		Title:     noteEntity.Title,
		Body:      noteEntity.Body,
		FolderID:  noteEntity.FolderID,
		Folder:    folderModel,
		Shared:    shared,
		CreatedAt: noteEntity.CreatedAt,
		UpdatedAt: noteEntity.UpdatedAt,
	}
	return m
}
