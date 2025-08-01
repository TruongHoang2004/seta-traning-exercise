package controllers

import (
	"collab-service/internal/database"
	"collab-service/internal/dto"
	"collab-service/internal/middleware"
	"collab-service/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShareFolder godoc
// @Summary Share folder with user
// @Description Share a folder with another user with read or write access
// @Tags folders
// @Accept json
// @Produce json
// @Param folderId path string true "Folder ID"
// @Param body body dto.ShareDTO true "User ID and permission"
// @Success 200 {object} map[string]interface{} "Success response with message"
// @Failure 400,401,403,404,500 {object} map[string]interface{} "Error response"
// @Router /folders/{folderId}/share [post]
func ShareFolder(c *gin.Context) {
	var body dto.ShareDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folderId := c.Param("folderId")
	userId, _ := middleware.GetUserInfoFromGin(c)

	// Check ownership
	var folder models.Folder
	if err := database.DB.Where("id = ? AND owner_id = ?", folderId, userId).First(&folder).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to share this folder"})
		return
	}

	// Create share
	share := models.FolderShare{
		FolderID: folder.ID,
		UserID:   body.UserID,
		Access:   body.AccessRole,
	}
	if err := database.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

// RevokeFolderShare godoc
// @Summary Revoke shared folder access
// @Description Remove a user's access to a shared folder
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{} "Success response with message"
// @Failure 403,404,500 {object} map[string]interface{} "Error response"
// @Router /folders/{folderId}/share/{userId} [delete]
func RevokeFolderShare(c *gin.Context) {
	folderId := c.Param("folderId")
	sharedUserId := c.Param("userId")
	userId, _ := middleware.GetUserInfoFromGin(c)

	var folder models.Folder
	if err := database.DB.Where("id = ? AND owner_id = ?", folderId, userId).First(&folder).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to revoke sharing of this folder"})
		return
	}

	if err := database.DB.Where("folder_id = ? AND user_id = ?", folderId, sharedUserId).Delete(&models.FolderShare{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder sharing revoked"})
}

// ShareNote godoc
// @Summary Share a note with another user
// @Description Share a single note with read or write access
// @Tags notes
// @Accept json
// @Produce json
// @Param noteId path string true "Note ID"
// @Param body body dto.ShareDTO true "User ID and permission"
// @Success 200 {object} map[string]interface{} "Success response with message"
// @Failure 403,404,500 {object} map[string]interface{} "Error response"
// @Router /notes/{noteId}/share [post]
func ShareNote(c *gin.Context) {
	var body dto.ShareDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	if body.AccessRole == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access role is required"})
		return
	}

	noteId := c.Param("noteId")
	userId, _ := middleware.GetUserInfoFromGin(c)

	var note models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", noteId, userId).First(&note).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to share this note"})
		return
	}

	share := models.NoteShare{
		NoteID: note.ID,
		UserID: body.UserID,
		Access: body.AccessRole,
	}
	if err := database.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note shared successfully"})
}

// RevokeNoteShare godoc
// @Summary Revoke note sharing
// @Description Remove a user's access to a shared note
// @Tags notes
// @Produce json
// @Param noteId path string true "Note ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{} "Success response with message"
// @Failure 403,404,500 {object} map[string]interface{} "Error response"
// @Router /notes/{noteId}/share/{userId} [delete]
func RevokeNoteShare(c *gin.Context) {
	noteId := c.Param("noteId")
	sharedUserId := c.Param("userId")
	userId, _ := middleware.GetUserInfoFromGin(c)

	var note models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", noteId, userId).First(&note).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to revoke sharing of this note"})
		return
	}

	if err := database.DB.Where("note_id = ? AND user_id = ?", noteId, sharedUserId).Delete(&models.NoteShare{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note sharing revoked"})
}
