package cache

import (
	"collab-service/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type NoteRepositoryWithCache struct {
	dbRepo entity.NoteRepository
	rdb    *redis.Client
	ttl    time.Duration
}

func NewNoteRepositoryWithCache(dbRepo entity.NoteRepository, rdb *redis.Client, ttl time.Duration) entity.NoteRepository {
	return &NoteRepositoryWithCache{
		dbRepo: dbRepo,
		rdb:    rdb,
		ttl:    ttl,
	}
}

// ChangeAccessLevel implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) ChangeAccessLevel(ctx context.Context, userID uuid.UUID, folderID uuid.UUID, accessLevel entity.AccessLevel) error {
	return n.dbRepo.ChangeAccessLevel(ctx, userID, folderID, accessLevel)
}

// Create implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) Create(ctx context.Context, note *entity.Note, userID uuid.UUID) (*entity.Note, error) {
	return n.dbRepo.Create(ctx, note, userID)
}

// GetOwner implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) GetOwner(ctx context.Context, noteID uuid.UUID) (uuid.UUID, error) {
	return n.dbRepo.GetOwner(ctx, noteID)
}

// Delete implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) Delete(ctx context.Context, id uuid.UUID) error {
	return n.dbRepo.Delete(ctx, id)
}

// GetAccessLevel implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) GetAccessLevel(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (entity.AccessLevel, error) {
	return n.dbRepo.GetAccessLevel(ctx, noteID, userID)
}

// GetAllCanAccess implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) GetAllCanAccess(ctx context.Context, userID uuid.UUID) ([]*entity.Note, error) {
	return n.dbRepo.GetAllCanAccess(ctx, userID)
}

// GetByFolderID implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) GetByFolderID(ctx context.Context, folderID uuid.UUID) ([]*entity.Note, error) {
	return n.dbRepo.GetByFolderID(ctx, folderID)
}

// GetByID implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) GetByID(ctx context.Context, id uuid.UUID) (*entity.Note, error) {
	return n.dbRepo.GetByID(ctx, id)
}

// GetFolderAccessLevel implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) GetFolderAccessLevel(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) (entity.AccessLevel, error) {
	return n.dbRepo.GetFolderAccessLevel(ctx, folderID, userID)
}

// RevokeAccess implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) RevokeAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
	return n.dbRepo.RevokeAccess(ctx, noteID, userID)
}

// ShareNote implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) ShareNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return n.dbRepo.ShareNote(ctx, noteID, userID, accessLevel)
}

// Update implements entity.NoteRepository.
func (n *NoteRepositoryWithCache) Update(ctx context.Context, note *entity.Note) error {
	return n.dbRepo.Update(ctx, note)
}
