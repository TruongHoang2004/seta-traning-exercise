package cache

import (
	"collab-service/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type FolderRepositoryWithCache struct {
	repo entity.FolderRepository
	rdb  *redis.Client
	ttl  time.Duration
}

func NewFolderRepositoryWithCache(repo entity.FolderRepository, rdb *redis.Client, ttl time.Duration) entity.FolderRepository {
	return &FolderRepositoryWithCache{
		repo: repo,
		rdb:  rdb,
		ttl:  ttl,
	}
}

// ChangeAccessLevel implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) ChangeAccessLevel(ctx context.Context, folderID uuid.UUID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return f.repo.ChangeAccessLevel(ctx, folderID, userID, accessLevel)
}

// Create implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) Create(ctx context.Context, folder *entity.Folder) (*entity.Folder, error) {
	return f.repo.Create(ctx, folder)
}

// Delete implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) Delete(ctx context.Context, id uuid.UUID) error {
	return f.repo.Delete(ctx, id)
}

// GetAccessLevel implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) GetAccessLevel(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) (entity.AccessLevel, error) {
	return f.repo.GetAccessLevel(ctx, folderID, userID)
}

// GetAllForCanAccess implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) GetAllForCanAccess(ctx context.Context, userID uuid.UUID) ([]*entity.Folder, error) {
	return f.repo.GetAllForCanAccess(ctx, userID)
}

// GetByID implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) GetByID(ctx context.Context, id uuid.UUID) (*entity.Folder, error) {
	return f.repo.GetByID(ctx, id)
}

// GetOwner implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) GetOwner(ctx context.Context, folderID uuid.UUID) (uuid.UUID, error) {
	return f.repo.GetOwner(ctx, folderID)
}

// RevokeAccess implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) RevokeAccess(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) error {
	return f.repo.RevokeAccess(ctx, folderID, userID)
}

// ShareFolder implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) ShareFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	return f.repo.ShareFolder(ctx, folderID, userID, accessLevel)
}

// Update implements entity.FolderRepository.
func (f *FolderRepositoryWithCache) Update(ctx context.Context, folder *entity.Folder) error {
	return f.repo.Update(ctx, folder)
}
