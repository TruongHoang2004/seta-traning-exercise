package persistence

import (
	"collab-service/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

type NoteRepositoryImpl struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) entity.NoteRepository {
	return &NoteRepositoryImpl{
		db: db,
	}
}

func (r *NoteRepositoryImpl) Create(ctx context.Context, note *entity.Note, userId uuid.UUID) error {
	model := NoteModelFromDomain(note)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the note
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		// Create a note share record giving the user OWNER access
		noteShare := NoteShareModel{
			NoteID:      model.ID,
			UserID:      userId,
			AccessLevel: entity.AccessLevelOwner,
		}

		if err := tx.Create(&noteShare).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *NoteRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Note, error) {
	var model NoteModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *NoteRepositoryImpl) GetFolderAccessLevel(ctx context.Context, folderID, userID uuid.UUID) (entity.AccessLevel, error) {
	var accessLevel entity.AccessLevel
	err := r.db.WithContext(ctx).
		Table("folders").
		Select("access_level").
		Where("id = ? AND user_id = ?", folderID, userID).
		Scan(&accessLevel).Error
	if err != nil {
		return entity.AccessLevelNone, err
	}
	return accessLevel, nil
}

// GetAccessLevel implements entity.NoteRepository.
func (r *NoteRepositoryImpl) GetAccessLevel(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (entity.AccessLevel, error) {
	type result struct {
		NoteAccessLevel   entity.AccessLevel
		FolderAccessLevel entity.AccessLevel
	}

	var rs result

	err := r.db.WithContext(ctx).
		Table("notes").
		Select(`
			note_shares.access_level AS note_access_level,
			COALESCE(folder_shares.access_level, CASE WHEN folders.user_id = ? THEN 'OWNER' END) AS folder_access_level
		`, userID).
		Joins("LEFT JOIN note_shares ON notes.id = note_shares.note_id AND note_shares.user_id = ?", userID).
		Joins("LEFT JOIN folders ON notes.folder_id = folders.id").
		Joins("LEFT JOIN folder_shares ON folders.id = folder_shares.folder_id AND folder_shares.user_id = ?", userID).
		Where("notes.id = ?", noteID).
		Scan(&rs).Error
	if err != nil {
		return entity.AccessLevelNone, err
	}

	// Lấy quyền cao nhất giữa note (nếu share trực tiếp) và folder (nếu là owner hoặc share)
	return entity.MaxAccessLevel(rs.NoteAccessLevel, rs.FolderAccessLevel), nil
}

// GetByFolderID implements entity.NoteRepository.
func (r *NoteRepositoryImpl) GetByFolderID(ctx context.Context, folderID uuid.UUID) ([]*entity.Note, error) {
	var models []NoteModel
	if err := r.db.WithContext(ctx).Model(&NoteModel{}).Where("folder_id = ?", folderID).Find(&models).Error; err != nil {
		return nil, err
	}

	notes := make([]*entity.Note, len(models))
	for i, model := range models {
		notes[i] = model.ToDomain()
	}
	return notes, nil
}

func (r *NoteRepositoryImpl) GetAllCanAccess(ctx context.Context, userID uuid.UUID) ([]*entity.Note, error) {
	var models []NoteModel
	if err := r.db.WithContext(ctx).Model(&NoteModel{}).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}

	// Lấy tất cả notes mà user có quyền:
	//  1) Share trực tiếp qua note_shares
	//  2) Nằm trong các folder mà user là owner hoặc được share qua folders_share
	if err := r.db.WithContext(ctx).
		Model(&NoteModel{}).
		Joins("LEFT JOIN folders ON notes.folder_id = folders.id").
		Joins("LEFT JOIN note_shares ON notes.id = note_shares.note_id").
		Joins("LEFT JOIN folders_share ON folders.id = folders_share.folder_id").
		Where(`
        folders.user_id = ?                    -- user sở hữu folder chứa note
        OR note_shares.user_id = ?            -- note được share trực tiếp
        OR folders_share.user_id = ?          -- folder được share với user
    `, userID, userID, userID).
		Group("notes.id"). // tránh duplicate
		Find(&models).Error; err != nil {
		return nil, err
	}

	notes := make([]*entity.Note, len(models))
	for i, model := range models {
		notes[i] = model.ToDomain()
	}
	return notes, nil
}

func (r *NoteRepositoryImpl) ShareNote(ctx context.Context, noteID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return r.db.WithContext(ctx).Table("note_shares").Create(map[string]interface{}{
		"note_id":      noteID,
		"user_id":      userID,
		"access_level": accessLevel,
	}).Error
}

func (r *NoteRepositoryImpl) RevokeAccess(ctx context.Context, noteID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("note_shares").Where("note_id = ? AND user_id = ?", noteID, userID).Delete(nil).Error
}

func (r *NoteRepositoryImpl) ChangeAccessLevel(ctx context.Context, noteID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return r.db.WithContext(ctx).Table("note_shares").
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Update("access_level", accessLevel).Error
}

func (r *NoteRepositoryImpl) Update(ctx context.Context, note *entity.Note) error {
	model := NoteModelFromDomain(note)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *NoteRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&NoteModel{}, "id = ?", id).Error
}
