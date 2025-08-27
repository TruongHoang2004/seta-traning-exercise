package persistence

import (
	"collab-service/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

type FolderRepositoryImpl struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) entity.FolderRepository {
	return &FolderRepositoryImpl{
		db: db,
	}
}

// Create implements entity.FolderRepository.
func (f *FolderRepositoryImpl) Create(ctx context.Context, folder *entity.Folder) (*entity.Folder, error) {
	folderModel := FolderModelFromDomain(folder)
	if err := f.db.WithContext(ctx).Create(folderModel).Error; err != nil {
		return &entity.Folder{}, err
	}
	return folderModel.ToDomain(), nil
}

// GetOwner implements entity.FolderRepository.
func (f *FolderRepositoryImpl) GetOwner(ctx context.Context, folderID uuid.UUID) (uuid.UUID, error) {
	var ownerID string
	if err := f.db.WithContext(ctx).Table("folder_shares").Where("folder_id = ? AND access_level = ?", folderID, entity.AccessLevelOwner).Select("user_id").Scan(&ownerID).Error; err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(ownerID)
}

func (f *FolderRepositoryImpl) Delete(ctx context.Context, folderID uuid.UUID) error {
	return f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Xóa folder_shares
		if err := tx.Where("folder_id = ?", folderID).Delete(&FolderShareModel{}).Error; err != nil {
			return err
		}

		// 2. Lấy danh sách note_id trong folder
		var noteIDs []uuid.UUID
		if err := tx.Model(&NoteModel{}).
			Where("folder_id = ?", folderID).
			Pluck("id", &noteIDs).Error; err != nil {
			return err
		}

		if len(noteIDs) > 0 {
			// 3. Xóa note_shares của các note này
			if err := tx.Where("note_id IN ?", noteIDs).Delete(&NoteShareModel{}).Error; err != nil {
				return err
			}

			// 4. Xóa notes
			if err := tx.Where("id IN ?", noteIDs).Delete(&NoteModel{}).Error; err != nil {
				return err
			}
		}

		// 5. Xóa folder
		if err := tx.Delete(&FolderModel{}, "id = ?", folderID).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetAccessLevel implements entity.FolderRepository.
func (f *FolderRepositoryImpl) GetAccessLevel(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) (entity.AccessLevel, error) {
	var accessLevelStr string
	var shareModel FolderShareModel

	if err := f.db.WithContext(ctx).Table("folder_shares").Where("folder_id = ? AND user_id = ?", folderID, userID).First(&shareModel).Error; err != nil {
		return entity.AccessLevelNone, err
	}

	accessLevelStr = string(shareModel.AccessLevel)

	return entity.AccessLevel(accessLevelStr), nil

}

// GetAllForCanAccess implements entity.FolderRepository.
func (f *FolderRepositoryImpl) GetAllForCanAccess(ctx context.Context, userID uuid.UUID) ([]*entity.Folder, error) {
	var models []FolderModel
	if err := f.db.WithContext(ctx).Model(&FolderModel{}).Joins("JOIN folder_shares ON folder_shares.folder_id = folders.id").Where("folder_shares.user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}

	folders := make([]*entity.Folder, len(models))
	for i, model := range models {
		folders[i] = model.ToDomain()
	}
	return folders, nil
}

// GetByID implements entity.FolderRepository.
func (f *FolderRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Folder, error) {
	var model FolderModel
	if err := f.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// Update implements entity.FolderRepository.
func (f *FolderRepositoryImpl) Update(ctx context.Context, folder *entity.Folder) error {
	model := FolderModelFromDomain(folder)
	return f.db.WithContext(ctx).Model(&FolderModel{ID: folder.ID}).Updates(map[string]interface{}{
		"name": model.Name,
	}).Error
}

func (r *FolderRepositoryImpl) ShareFolder(ctx context.Context, folderID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return r.db.WithContext(ctx).Table("folder_shares").Create(map[string]interface{}{
		"folder_id":    folderID,
		"user_id":      userID,
		"access_level": accessLevel,
	}).Error
}

func (r *FolderRepositoryImpl) RevokeAccess(ctx context.Context, folderID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("folder_shares").Where("folder_id = ? AND user_id = ?", folderID, userID).Delete(nil).Error
}

func (r *FolderRepositoryImpl) ChangeAccessLevel(ctx context.Context, folderID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return r.db.WithContext(ctx).Table("folder_shares").
		Where("folder_id = ? AND user_id = ?", folderID, userID).
		Update("access_level", accessLevel).Error
}
