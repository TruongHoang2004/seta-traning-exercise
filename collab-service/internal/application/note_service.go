package application

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/event"
	"collab-service/internal/interface/http/middleware"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NoteService struct {
	repo          entity.NoteRepository
	eventProducer *event.AssetChangeProducer
}

func NewNoteService(repo entity.NoteRepository) *NoteService {
	return &NoteService{
		repo:          repo,
		eventProducer: event.GetAssetChangeProducer(),
	}
}

func (s *NoteService) Create(c *gin.Context, note *entity.Note) (*entity.Note, error) {
	// Get user ID from the context
	userID, _ := middleware.GetUserInfoFromGin(c)

	ownerID, err := s.repo.GetOwner(c.Request.Context(), note.FolderID)
	if err != nil {
		return nil, err
	}

	// If the note has a folder, check folder access
	if note.FolderID != uuid.Nil {
		folderAccessLevel, _ := s.repo.GetFolderAccessLevel(c.Request.Context(), note.FolderID, userID)
		if !folderAccessLevel.GreaterThan(entity.AccessLevelWrite) {
			return nil, NewForbiddenError("You do not have write permission for this folder")
		}
	}

	savedNote, err := s.repo.Create(c.Request.Context(), note, userID)

	go s.eventProducer.Produce(event.NewAssetEvent(event.NoteCreated, event.Note, savedNote.ID.String(), ownerID.String(), userID.String(), time.Now().String()))

	return savedNote, err
}

func (s *NoteService) GetByID(c *gin.Context, id uuid.UUID) (*entity.Note, error) {
	return s.repo.GetByID(c.Request.Context(), id)
}

func (s *NoteService) GetByFolderID(c *gin.Context, folderID uuid.UUID) ([]*entity.Note, error) {
	return s.repo.GetByFolderID(c.Request.Context(), folderID)
}

func (s *NoteService) GetAllCanAccess(c *gin.Context, userID uuid.UUID) ([]*entity.Note, error) {
	return s.repo.GetAllCanAccess(c.Request.Context(), userID)
}

func (s *NoteService) GetAccessLevel(c *gin.Context, noteID, userID uuid.UUID) (entity.AccessLevel, error) {
	return s.repo.GetAccessLevel(c.Request.Context(), noteID, userID)
}

func (s *NoteService) ShareNote(c *gin.Context, noteID, userID uuid.UUID, accessLevel entity.AccessLevel) error {
	// Check if the user has permission to share the note
	currentUserID, _ := middleware.GetUserInfoFromGin(c)

	currentAccessLevel, _ := s.GetAccessLevel(c, noteID, currentUserID)
	if !currentAccessLevel.GreaterThan(entity.AccessLevelOwner) {
		return NewForbiddenError("You do not have permission to share this note")
	}

	userAccessLevel, _ := s.GetAccessLevel(c, noteID, userID)

	if userAccessLevel != entity.AccessLevelNone {
		return s.repo.ChangeAccessLevel(c.Request.Context(), noteID, userID, accessLevel)
	}

	err := s.repo.ShareNote(c.Request.Context(), noteID, userID, accessLevel)

	go s.eventProducer.Produce(event.NewAssetEvent(event.NoteShared, event.Note, noteID.String(), userID.String(), currentUserID.String(), time.Now().String()))

	return err
}

func (s *NoteService) RevokeAccess(c *gin.Context, noteID, userID uuid.UUID) error {
	// Check if the user has permission to revoke access
	currentUserID, _ := middleware.GetUserInfoFromGin(c)
	currentAccessLevel, _ := s.GetAccessLevel(c, noteID, currentUserID)
	if !currentAccessLevel.GreaterThan(entity.AccessLevelOwner) {
		return NewForbiddenError("You do not have permission to revoke access for this note")
	}

	err := s.repo.RevokeAccess(c.Request.Context(), noteID, userID)

	go s.eventProducer.Produce(event.NewAssetEvent(event.NoteUnshared, event.Note, noteID.String(), userID.String(), currentUserID.String(), time.Now().String()))

	return err
}

func (s *NoteService) Update(c *gin.Context, note *entity.Note) error {
	userID, _ := middleware.GetUserInfoFromGin(c)

	var accessLevel entity.AccessLevel
	accessLevel, _ = s.GetAccessLevel(c, note.ID, userID)

	log.Println(accessLevel)

	if !accessLevel.GreaterThan(entity.AccessLevelWrite) {
		return NewForbiddenError("You do not have permission to update this note")
	}

	err := s.repo.Update(c.Request.Context(), note)
	go s.eventProducer.Produce(event.NewAssetEvent(event.NoteUpdated, event.Note, note.ID.String(), userID.String(), userID.String(), time.Now().String()))
	return err
}

func (s *NoteService) Delete(c *gin.Context, id uuid.UUID) error {
	userID, _ := middleware.GetUserInfoFromGin(c)

	var accessLevel entity.AccessLevel
	accessLevel, _ = s.GetAccessLevel(c, id, userID)

	if !accessLevel.GreaterThan(entity.AccessLevelOwner) {
		return NewForbiddenError("You do not have permission to delete this note")
	}

	err := s.repo.Delete(c.Request.Context(), id)

	go s.eventProducer.Produce(event.NewAssetEvent(event.NoteDeleted, event.Note, id.String(), userID.String(), userID.String(), time.Now().String()))

	return err
}
