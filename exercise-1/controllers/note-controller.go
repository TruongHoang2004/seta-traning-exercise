package controllers

import (
	"fmt"
	"net/http"
	"seta-training-exercise-1/database"
	"seta-training-exercise-1/dto"
	"seta-training-exercise-1/middleware"
	"seta-training-exercise-1/models"

	"github.com/gin-gonic/gin"
)

// CreateNote creates a new note
// @Security BearerAuth
// @Summary Create a new note
// @Description Create a new note in a folder
// @Tags notes
// @Accept json
// @Produce json
// @Param note body dto.NoteDTO true "Note data"
// @Success 201 {object} models.Note "Created note"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 404 {object} object "User not found"
// @Router /notes [post]
func CreateNote(c *gin.Context) {
	var noteDTO dto.NoteDTO

	if err := c.ShouldBindJSON(&noteDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	result := database.DB.First(&user, "id = ?", userId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	note := models.Note{
		Title:    noteDTO.Title,
		Body:     noteDTO.Body,
		FolderID: noteDTO.FolderID,
		OwnerID:  user.ID,
	}

	if err := database.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote returns note details with read permission
// @Security BearerAuth
// @Summary Get a note
// @Tags notes
// @Produce json
// @Param noteId path string true "Note ID"
// @Success 200 {object} models.Note
// @Failure 401 {object} object
// @Failure 403 {object} object
// @Failure 404 {object} object
// @Router /notes/{noteId} [get]
func GetNote(c *gin.Context) {
	noteId := c.Param("noteId")
	userId, _ := middleware.GetUserIDFromGin(c)

	type NoteWithAccess struct {
		models.Note
		Access     *string
		UserExists *string
	}

	var note NoteWithAccess

	result := database.DB.
		Table("notes").
		Select("notes.*, note_shares.access as access, users.id as user_exists").
		Joins("LEFT JOIN note_shares ON notes.id = note_shares.note_id AND note_shares.user_id = ?", userId).
		Joins("LEFT JOIN users ON users.id = ?", userId).
		Where("notes.id = ?", noteId).
		First(&note)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.UserExists == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if note.OwnerID != userId {
		if note.Access == nil || (*note.Access != string(models.AccessLevelRead) && *note.Access != string(models.AccessLevelWrite)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this note"})
			return
		}
	}

	c.JSON(http.StatusOK, note.Note)
}

// UpdateNote updates a note
// @Security BearerAuth
// @Summary Update a note
// @Tags notes
// @Accept json
// @Produce json
// @Param noteId path string true "Note ID"
// @Param note body dto.NoteDTO true "Updated note"
// @Success 200 {object} models.Note
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 403 {object} object
// @Failure 404 {object} object
// @Router /notes/{noteId} [put]
func UpdateNote(c *gin.Context) {
	noteId := c.Param("noteId")
	userId, _ := middleware.GetUserIDFromGin(c)

	var noteDTO dto.NoteDTO
	if err := c.ShouldBindJSON(&noteDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type NoteWithAccess struct {
		models.Note
		Access     *string
		UserExists *string
	}

	var note NoteWithAccess

	result := database.DB.
		Table("notes").
		Select("notes.*, note_shares.access as access, users.id as user_exists").
		Joins("LEFT JOIN note_shares ON notes.id = note_shares.note_id AND note_shares.user_id = ?", userId).
		Joins("LEFT JOIN users ON users.id = ?", userId).
		Where("notes.id = ?", noteId).
		First(&note)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.UserExists == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if note.OwnerID != userId {
		if note.Access == nil || *note.Access != string(models.AccessLevelWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have WRITE permission on this note"})
			return
		}
	}

	note.Title = noteDTO.Title
	note.Body = noteDTO.Body

	if err := database.DB.Save(&note.Note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note.Note)
}

// DeleteNote deletes a note and all shares
// @Security BearerAuth
// @Summary Delete a note
// @Tags notes
// @Produce json
// @Param noteId path string true "Note ID"
// @Success 200 {object} object
// @Failure 401 {object} object
// @Failure 403 {object} object
// @Failure 404 {object} object
// @Router /notes/{noteId} [delete]
func DeleteNote(c *gin.Context) {
	noteId := c.Param("noteId")
	userId, _ := middleware.GetUserIDFromGin(c)

	var note models.Note
	result := database.DB.First(&note, "id = ?", noteId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this note"})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Where("note_id = ?", noteId).Delete(&models.NoteShare{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note shares"})
		return
	}

	if err := tx.Delete(&note).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Note %s deleted successfully", noteId)})
}
