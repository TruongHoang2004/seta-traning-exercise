package handler

import (
	"collab-service/internal/application"
	"collab-service/internal/domain/entity"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// @Security BearerAuth
// ImportUsersFromCSV imports multiple users from a CSV file
// @Summary Import users from CSV
// @Description Import multiple users from a CSV file
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file containing user data"
// @Router /users/import [post]
func (h *UserHandler) ImportUsersFromCSV(c *gin.Context) {
	// Get the uploaded file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get CSV file: %v", err)})
		return
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to open CSV file: %v", err)})
		return
	}
	defer file.Close()

	// Create CSV reader
	csvReader := csv.NewReader(file)

	// Skip header row
	if _, err := csvReader.Read(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to read CSV header: %v", err)})
		return
	}

	var users []*entity.User

	records, err := csvReader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to read CSV file: %v", err)})
		return
	}

	for i, record := range records {
		if len(record) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("line %d: not enough fields", i+2)})
			return
		}

		user := &entity.User{
			Username: record[0],
			Email:    record[1],
			Password: record[2],
			Role:     entity.UserType(record[3]),
		}
		users = append(users, user)
	}

	createdUsers, errs := h.userService.CreateManyUsers(c.Request.Context(), users)
	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create users: %v", errs)})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"message": "users imported successfully",
		"users":   createdUsers,
	})
}
