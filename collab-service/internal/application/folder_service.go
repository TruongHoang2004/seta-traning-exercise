package application

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/event"
	"collab-service/internal/interface/http/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FolderService struct {
	folderRepo    entity.FolderRepository
	eventProducer *event.AssetChangeProducer
}

func NewFolderService(repo entity.FolderRepository) *FolderService {
	return &FolderService{
		folderRepo:    repo,
		eventProducer: event.GetAssetChangeProducer(),
	}
}

func (s *FolderService) Create(c *gin.Context, name string) (*entity.Folder, error) {
	folder := entity.NewFolder(name)

	userId, _ := middleware.GetUserInfoFromGin(c)
	folder.Shared = append(folder.Shared, entity.FolderShare{
		UserID:      userId,
		AccessLevel: entity.AccessLevelOwner,
	})

	createdFolder, err := s.folderRepo.Create(c.Request.Context(), folder)
	if err != nil {
		return nil, NewBadRequestError(err.Error())
	}

	go func() {
		event := event.NewAssetEvent(
			event.FolderCreated,
			event.Folder,
			createdFolder.ID.String(),
			userId.String(),
			userId.String(),
			time.Now().String(),
		)
		if err := s.eventProducer.Produce(event); err != nil {
			// Handle error
		}
	}()
	return createdFolder, nil
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

	err := s.folderRepo.ShareFolder(c.Request.Context(), folderID, userID, accessLevel)

	go s.eventProducer.Produce(event.NewAssetEvent(
		event.FolderShared,
		event.Folder,
		folderID.String(),
		userID.String(),
		currentUserID.String(),
		time.Now().String(),
	))

	return err
}

func (s *FolderService) RevokeAccess(c *gin.Context, folderID, userID uuid.UUID) error {
	// Check if the user has permission to revoke access
	currentUserID, _ := middleware.GetUserInfoFromGin(c)
	currentAccessLevel, _ := s.folderRepo.GetAccessLevel(c.Request.Context(), folderID, currentUserID)
	if !currentAccessLevel.GreaterThan(entity.AccessLevelOwner) {
		return NewForbiddenError("You do not have permission to revoke access for this folder")
	}

	err := s.folderRepo.RevokeAccess(c.Request.Context(), folderID, userID)

	go s.eventProducer.Produce(event.NewAssetEvent(
		event.FolderUnshared,
		event.Folder,
		folderID.String(),
		currentUserID.String(),
		userID.String(),
		time.Now().String(),
	))

	return err
}

func (s *FolderService) Update(c *gin.Context, folder *entity.Folder) error {

	userID, _ := middleware.GetUserInfoFromGin(c)

	role, err := s.folderRepo.GetAccessLevel(c.Request.Context(), folder.ID, userID)
	if err != nil {
		return NewNotFoundError(err.Error())
	}

	if !role.GreaterThan(entity.AccessLevelWrite) {
		return NewForbiddenError("You do not have permission to update this folder")
	}

	if err := s.folderRepo.Update(c.Request.Context(), folder); err != nil {
		return NewBadRequestError(err.Error())
	}

	ownerId, err := s.folderRepo.GetOwner(c.Request.Context(), folder.ID)
	if err != nil {
		return NewNotFoundError(err.Error())
	}

	go s.eventProducer.Produce(event.NewAssetEvent(
		event.FolderUpdated,
		event.Folder,
		folder.ID.String(),
		ownerId.String(),
		userID.String(),
		time.Now().String(),
	))

	return nil
}

func (s *FolderService) Delete(c *gin.Context, id uuid.UUID) error {

	userID, _ := middleware.GetUserInfoFromGin(c)

	ownerID, err := s.folderRepo.GetOwner(c.Request.Context(), id)
	if err != nil {
		return NewNotFoundError(err.Error())
	}

	if userID != ownerID {
		return NewForbiddenError("You do not have permission to delete this folder")
	}

	if err := s.folderRepo.Delete(c.Request.Context(), id); err != nil {
		return NewBadRequestError(err.Error())
	}

	go func() {
		event := event.NewAssetEvent(
			event.FolderDeleted,
			event.Folder,
			id.String(),
			ownerID.String(),
			userID.String(),
			time.Now().String(),
		)
		if err := s.eventProducer.Produce(event); err != nil {
			// Handle error
		}
	}()

	return nil
}
