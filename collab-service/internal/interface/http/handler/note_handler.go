package handler

import (
	"collab-service/internal/application"
	"collab-service/internal/domain/entity"
	"collab-service/internal/interface/http/dto"
	"collab-service/internal/interface/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NoteHandler struct {
	NoteService *application.NoteService
}

func NewNoteHandler(noteService *application.NoteService) *NoteHandler {
	return &NoteHandler{
		NoteService: noteService,
	}
}

// @Security BearerAuth
// Create godoc
// @Summary Create a new note
// @Description Creates a new note with the given details
// @Tags notes
// @Accept json
// @Produce json
// @Param body body dto.CreateNoteRequest true "Note details"
// @Router /notes [post]
func (h *NoteHandler) Create(c *gin.Context) {
	var req dto.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note := &entity.Note{
		Title:    req.Title,
		Body:     req.Body,
		FolderID: req.FolderID,
	}

	if err := h.NoteService.Create(c, note); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, note)
}

// @Security BearerAuth
// @Summary Get a note by ID
// @Description Get a note by its unique ID
// @Tags notes
// @Produce json
// @Param id path string true "Note ID"
// @Router /notes/{id} [get]
func (h *NoteHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	note, err := h.NoteService.GetByID(c, parsedID)
	if err != nil {
		application.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, note)
}

// GetAll godoc
// @Security BearerAuth
// @Summary Get all notes
// @Description Get all notes that the user has access to
// @Tags notes
// @Produce json
// @Router /notes [get]
func (h *NoteHandler) GetAll(c *gin.Context) {
	userID, _ := middleware.GetUserInfoFromGin(c)

	notes, err := h.NoteService.GetAllCanAccess(c, userID)
	if err != nil {
		application.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, notes)
}

// @Security BearerAuth
// @Summary Share a note with another user
// @Description Share a note with another user by ID and access level
// @Tags notes
// @Accept json
// @Produce json
// @Param id path string true "Note ID"
// @Param body body dto.ShareNoteRequest true "Share note details"
// @Router /notes/{noteID}/share [post]
func (h *NoteHandler) ShareNote(c *gin.Context) {
	var req dto.ShareNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	if err := h.NoteService.ShareNote(c, noteID, req.UserID, req.AccessLevel); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// @Security BearerAuth
// @Summary Revoke access to a note
// @Description Revoke access to a note for a specific user
// @Tags notes
// @Router /notes/{noteId}/share/{userId} [delete]
func (h *NoteHandler) RevokeAccess(c *gin.Context) {
	userID := c.Param("userID")
	noteID := c.Param("noteID")

	parsedNoteID, err := uuid.Parse(noteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.NoteService.RevokeAccess(c, parsedNoteID, parsedUserID); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// @Security BearerAuth
// @Summary Update a note
// @Description Update a note by its unique ID
// @Tags notes
// @Accept json
// @Produce json
// @Param id path string true "Note ID"
// @Param body body dto.UpdateNoteRequest true "Note details"
// @Router /notes/{id} [put]
func (h *NoteHandler) Update(c *gin.Context) {
	var req dto.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	note := &entity.Note{
		ID:       parsedID,
		Title:    req.Title,
		Body:     req.Body,
		FolderID: req.FolderID,
	}

	if err := h.NoteService.Update(c, note); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, note)
}

// @Security BearerAuth
// @Summary Delete a note
// @Description Delete a note by its unique ID
// @Tags notes
// @Produce json
// @Param id path string true "Note ID"
// @Router /notes/{id} [delete]
func (h *NoteHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.NoteService.Delete(c, parsedID); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
