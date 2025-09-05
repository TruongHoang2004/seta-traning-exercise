package model

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type NoteShareModel struct {
	ID     uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	NoteID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`

	AccessLevel entity.AccessLevel `gorm:"type:varchar(10);not null;default:'READ'"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (NoteShareModel) TableName() string {
	return "note_shares"
}

func (m *NoteShareModel) ToDomain() *entity.NoteShare {
	return &entity.NoteShare{
		ID:          m.ID,
		NoteID:      m.NoteID,
		UserID:      m.UserID,
		AccessLevel: m.AccessLevel,
		CreatedAt:   m.CreatedAt,
	}
}

func NoteShareModelFromDomain(noteShareEntity *entity.NoteShare) *NoteShareModel {
	return &NoteShareModel{
		ID:          noteShareEntity.ID,
		NoteID:      noteShareEntity.NoteID,
		UserID:      noteShareEntity.UserID,
		AccessLevel: noteShareEntity.AccessLevel,
		CreatedAt:   noteShareEntity.CreatedAt,
	}
}
