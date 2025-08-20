package application

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FolderService struct {
	folderRepo entity.FolderRepository
}

func NewFolderService(repo entity.FolderRepository) *FolderService {
	return &FolderService{
		folderRepo: repo,
	}
}

func (s *FolderService) Create(c *gin.Context, name string) (*entity.Folder, error) {
	folder := entity.NewFolder(name)

	if err := s.folderRepo.Create(c.Request.Context(), folder); err != nil {
		return nil, NewBadRequestError(err.Error())
	}
	return folder, nil
}

func (s *FolderService) GetFolderByID(c *gin.Context, id uuid.UUID) (*entity.Folder, error) {
	folder, err := s.folderRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		return nil, NewNotFoundError(err.Error())
	}
	return folder, nil
}

func (s *FolderService) GetAllFolderCanAccess(c *gin.Context) ([]*entity.Folder, error) {
	userId, _ := middleware.GetUserInfoFromGin(c)

	folders, err := s.folderRepo.GetAllForCanAccess(c.Request.Context(), userId)
	if err != nil {
		return nil, NewNotFoundError(err.Error())
	}
	return folders, nil
}

func (s *FolderService) ShareFolder(c *gin.Context, folderID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	// Check if the user has permission to share the folder
	currentUserID, _ := middleware.GetUserInfoFromGin(c)

	currentAccessLevel, _ := s.folderRepo.GetAccessLevel(c.Request.Context(), folderID, currentUserID)
	if !currentAccessLevel.GreaterThan(entity.AccessLevelOwner) {
		return NewForbiddenError("You do not have permission to share this folder")
	}

	userAccessLevel, _ := s.folderRepo.GetAccessLevel(c.Request.Context(), folderID, userID)

	if userAccessLevel != entity.AccessLevelNone {
		return s.folderRepo.ChangeAccessLevel(c.Request.Context(), folderID, userID, accessLevel)
	}

	return s.folderRepo.ShareFolder(c.Request.Context(), folderID, userID, accessLevel)
}

func (s *FolderService) RevokeAccess(c *gin.Context, folderID, userID uuid.UUID) error {
	// Check if the user has permission to revoke access
	currentUserID, _ := middleware.GetUserInfoFromGin(c)
	currentAccessLevel, _ := s.folderRepo.GetAccessLevel(c.Request.Context(), folderID, currentUserID)
	if !currentAccessLevel.GreaterThan(entity.AccessLevelOwner) {
		return NewForbiddenError("You do not have permission to revoke access for this folder")
	}

	return s.folderRepo.RevokeAccess(c.Request.Context(), folderID, userID)
}

func (s *FolderService) Update(c *gin.Context, folder *entity.Folder) error {
	if err := s.folderRepo.Update(c.Request.Context(), folder); err != nil {
		return NewBadRequestError(err.Error())
	}
	return nil
}

func (s *FolderService) Delete(c *gin.Context, id uuid.UUID) error {
	if err := s.folderRepo.Delete(c.Request.Context(), id); err != nil {
		return NewBadRequestError(err.Error())
	}
	return nil
}
