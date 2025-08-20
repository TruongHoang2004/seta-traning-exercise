package persistence

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type FolderShareModel struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FolderID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index"`

	AccessLevel entity.AccessLevel `gorm:"type:varchar(10);not null;default:'READ'"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (FolderShareModel) TableName() string {
	return "folder_shares"
}

func (m *FolderShareModel) ToDomain() *entity.FolderShare {
	return &entity.FolderShare{
		ID:          m.ID,
		FolderID:    m.FolderID,
		UserID:      m.UserID,
		AccessLevel: m.AccessLevel,
	}
}

func FolderShareModelFromDomain(folderShareEntity *entity.FolderShare) *FolderShareModel {
	return &FolderShareModel{
		ID:          folderShareEntity.ID,
		FolderID:    folderShareEntity.FolderID,
		UserID:      folderShareEntity.UserID,
		AccessLevel: folderShareEntity.AccessLevel,
	}
}
