package controllers

import (
	"collab-service/internal/database"
	"collab-service/internal/dto"
	"collab-service/internal/middleware"
	"collab-service/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateFolder creates a new folder
// @Security BearerAuth
// @Summary Create a new folder
// @Description Create a new folder for the authenticated user
// @Tags folders
// @Accept json
// @Produce json
// @Param folder body dto.FolderDTO true "Folder data"
// @Success 201 {object} models.Folder "Created folder"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 404 {object} object "User not found"
// @Failure 500 {object} object "Internal server error"
// @Router /folders [post]
func CreateFolder(c *gin.Context) {
	var folderDTO dto.FolderDTO

	if err := c.ShouldBindJSON(&folderDTO); err != nil {
		// custom error to response
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Kiểm tra folder trùng tên (tối ưu hóa bằng SELECT EXISTS)
	var exists bool
	// use gorm instead of raw SQL
	err = database.DB.
		Raw("SELECT EXISTS (SELECT 1 FROM folders WHERE owner_id = ? AND name = ?) AS exists", userId, folderDTO.Name).
		Scan(&exists).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing folder"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Folder with this name already exists"})
		return
	}

	folder := models.Folder{
		Name:    folderDTO.Name,
		OwnerID: userId,
	}

	// pass ctx to database.DB.Create
	if err := database.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// GetFolder retrieves folder details by ID
// @Security BearerAuth
// @Summary Get folder details
// @Description Retrieve details of a specific folder
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} models.Folder "Folder details"
// @Failure 401 {object} object "Unauthorized"
// @Failure 404 {object} object "Folder not found"
// @Router /folders/{folderId} [get]
func GetFolder(c *gin.Context) {
	folderId := c.Param("folderId")
	userId, _ := middleware.GetUserIDFromGin(c)

	type FolderWithAccess struct {
		models.Folder
		Access     *string // nullable vì user có thể không có chia sẻ
		UserExists *string // check nếu user tồn tại
	}

	var folder FolderWithAccess

	result := database.DB.
		Table("folders").
		Select("folders.*, folder_shares.access as access, users.id as user_exists").
		Joins("LEFT JOIN folder_shares ON folders.id = folder_shares.folder_id AND folder_shares.user_id = ?", userId).
		Joins("LEFT JOIN users ON users.id = ?", userId).
		Where("folders.id = ?", folderId).
		First(&folder)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if folder.UserExists == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if folder.OwnerID != userId {
		if folder.Access == nil || (*folder.Access != string(models.AccessLevelRead) && *folder.Access != string(models.AccessLevelWrite)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this folder"})
			return
		}
	}

	c.JSON(http.StatusOK, folder.Folder)

}

// UpdateFolder updates a folder's name and metadata
// @Security BearerAuth
// @Summary Update folder
// @Description Update a folder's name
// @Tags folders
// @Accept json
// @Produce json
// @Param folderId path string true "Folder ID"
// @Param folder body dto.FolderDTO true "Updated folder data"
// @Success 200 {object} models.Folder "Updated folder"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 404 {object} object "Folder not found"
// @Router /folders/{folderId} [put]
func UpdateFolder(c *gin.Context) {
	folderId := c.Param("folderId")
	userId, _ := middleware.GetUserIDFromGin(c)

	var folderDTO dto.FolderDTO
	if err := c.ShouldBindJSON(&folderDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type FolderWithAccess struct {
		models.Folder
		Access     *string // nullable: nếu không có chia sẻ
		UserExists *string
	}

	var folder FolderWithAccess

	result := database.DB.
		Table("folders").
		Select("folders.*, folder_shares.access as access, users.id as user_exists").
		Joins("LEFT JOIN folder_shares ON folders.id = folder_shares.folder_id AND folder_shares.user_id = ?", userId).
		Joins("LEFT JOIN users ON users.id = ?", userId).
		Where("folders.id = ?", folderId).
		First(&folder)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if folder.UserExists == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if folder.OwnerID != userId {
		if folder.Access == nil || *folder.Access != string(models.AccessLevelWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have WRITE permission on this folder"})
			return
		}
	}

	// Cập nhật tên folder
	folder.Name = folderDTO.Name
	if err := database.DB.Save(&folder.Folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, folder.Folder)
}

// DeleteFolder deletes a folder and its notes
// DeleteFolder deletes a folder and its notes
// @Security BearerAuth
// @Summary Delete folder
// @Description Delete a folder and all its notes
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} object "Success message"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Folder not found"
// @Router /folders/{folderId} [delete]
func DeleteFolder(c *gin.Context) {
	// add log when error
	folderId := c.Param("folderId")
	userId, _ := middleware.GetUserIDFromGin(c)

	// Truy vấn folder + xác thực quyền sở hữu
	var folder models.Folder
	result := database.DB.First(&folder, "id = ?", folderId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Chỉ cho phép owner xóa
	if folder.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this folder"})
		return
	}

	// Bắt đầu transaction để đảm bảo atomic
	tx := database.DB.Begin()

	// Xóa các ghi chú trong folder
	if err := tx.Where("folder_id = ?", folderId).Delete(&models.Note{}).Error; err != nil {
		// log.Error()
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notes"})
		return
	}

	// Xóa các quan hệ chia sẻ
	if err := tx.Where("folder_id = ?", folderId).Delete(&models.FolderShare{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder shares"})
		return
	}

	// Xóa folder
	if err := tx.Delete(&folder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Folder %s deleted successfully", folderId)})
}
