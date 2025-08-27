package handler

import (
	"collab-service/internal/application"
	"collab-service/internal/interface/http/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FolderHandler handles HTTP requests related to folders
type FolderHandler struct {
	folderService *application.FolderService
}

// NewFolderHandler creates a new instance of FolderHandler
func NewFolderHandler(folderService *application.FolderService) *FolderHandler {
	return &FolderHandler{
		folderService: folderService,
	}
}

// @Security BearerAuth
// @Summary Create a new folder
// @Description Create a new folder with the given name
// @Tags folders
// @Accept json
// @Produce json
// @Param request body dto.CreateFolderRequest true "Folder creation request"
// @Router /folders [post]
func (h *FolderHandler) Create(c *gin.Context) {
	var request dto.CreateFolderRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	folder, err := h.folderService.Create(c, request.Name)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response dto.FolderResponse
	response.ID = folder.ID
	response.Name = folder.Name

	c.JSON(http.StatusCreated, response)
}

// @Security BearerAuth
// @Summary Get a folder by ID
// @Description Get a folder by its unique ID
// @Tags folders
// @Produce json
// @Param id path string true "Folder ID"
// @Router /folders/{id} [get]
func (h *FolderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	folder, err := h.folderService.GetFolderByID(c, uuid.MustParse(id))
	if err != nil {
		application.HandleError(c, err)
		return
	}
	if folder == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	var response dto.FolderResponse
	response.ID = folder.ID
	response.Name = folder.Name

	c.JSON(http.StatusOK, response)
}

// @Security BearerAuth
// @Summary Get all folders that the user can access
// @Description Get all folders that the user has access to
// @Tags folders
// @Produce json
// @Router /folders [get]
func (h *FolderHandler) GetAllCanAccess(c *gin.Context) {
	folders, err := h.folderService.GetAllFolderCanAccess(c)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response []dto.FolderResponse
	for _, folder := range folders {
		response = append(response, dto.FolderResponse{
			ID:   folder.ID,
			Name: folder.Name,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Security BearerAuth
// @Summary Update a folder
// @Description Update a folder with the given ID and new name
// @Tags folders
// @Accept json
// @Produce json
// @Param folderID path string true "Folder ID"
// @Param request body dto.UpdateFolderRequest true "Folder update request"
// @Router /folders/{folderID} [put]
func (h *FolderHandler) Update(c *gin.Context) {
	var request dto.UpdateFolderRequest
	folderID, err := uuid.Parse(c.Param("folderID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	folder, err := h.folderService.GetFolderByID(c, folderID)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	folder.Name = request.Name

	if err := h.folderService.Update(c, folder); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder updated successfully"})
}

// @Security BearerAuth
// @Summary Delete a folder
// @Description Delete a folder by its unique ID
// @Tags folders
// @Produce json
// @Param folderID path string true "Folder ID"
// @Router /folders/{folderID} [delete]
func (h *FolderHandler) Delete(c *gin.Context) {
	folderID := c.Param("folderID")

	if err := h.folderService.Delete(c, uuid.MustParse(folderID)); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, "Success")
}

// @Security BearerAuth
// @Summary Share a folder with a user
// @Description Share a folder with a user by their ID and set access level
// @Tags folders
// @Accept json
// @Produce json
// @Param folderID path string true "Folder ID"
// @Param request body dto.ShareFolderRequest true "Folder sharing request"
// @Router /folders/{folderID}/share [post]
func (h *FolderHandler) ShareFolder(c *gin.Context) {
	var rep dto.ShareFolderRequest
	if err := c.ShouldBindJSON(&rep); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	folderID, err := uuid.Parse(c.Param("folderID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	if err := h.folderService.ShareFolder(c, folderID, rep.UserID, rep.AccessLevel); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

// @Security BearerAuth
// @Summary Unshare a folder with a user
// @Description Unshare a folder with a user by their ID
// @Tags folders
// @Produce json
// @Param folderID path string true "Folder ID"
// @Param userID path string true "User ID"
// @Router /folders/{folderID}/share/{userID} [delete]
func (h *FolderHandler) RevokeAccess(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	folderID, err := uuid.Parse(c.Param("folderID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	if err := h.folderService.RevokeAccess(c, folderID, userID); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder access revoked successfully"})
}
