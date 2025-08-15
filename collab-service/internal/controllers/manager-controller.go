package controllers

import (
	"collab-service/internal/database"
	"collab-service/internal/middleware"
	"collab-service/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTeamAssets godoc
// @Summary Get all accessible assets of team members
// @Description Manager-only API: View all assets that team members own or can access (shared folders or notes)
// @Tags assets
// @Produce json
// @Param teamId path string true "Team ID"
// @Success 200 {object} map[string]interface{} "Success response with owned and shared folders"
// @Failure 400,401,403,404,500 {object} map[string]interface{} "Error response"
// @Security BearerAuth
// @Router /teams/{teamId}/assets [get]
func GetTeamAssets(c *gin.Context) {
	teamID := c.Param("teamId")
	userID, _ := middleware.GetUserInfoFromGin(c) // middleware sets authenticated user

	db := database.DB.WithContext(c.Request.Context())

	if !isManagerOfTeam(c, userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team managers can access this"})
		return
	}

	// Lấy danh sách user trong team
	var rosters []models.Roster
	db.Where("team_id = ?", teamID).Preload("User").Find(&rosters)

	var userIDs []string
	for _, r := range rosters {
		userIDs = append(userIDs, r.UserID)
	}

	// Lấy folders do họ sở hữu hoặc được share
	var folders []models.Folder
	db.Where("owner_id IN ?", userIDs).Find(&folders)

	var sharedFolders []models.Folder
	db.Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id IN ?", userIDs).Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"owned_folders":  folders,
		"shared_folders": sharedFolders,
	})
}

// GetUserAssets godoc
// @Summary Get all assets of a user
// @Description Get all folders owned or shared with a specific user
// @Tags assets
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{} "Success response with owned and shared folders"
// @Failure 400,401,403,404,500 {object} map[string]interface{} "Error response"
// @Security BearerAuth
// @Router /users/{userId}/assets [get]
func GetUserAssets(c *gin.Context) {
	userID := c.Param("userId")
	authUserID, _ := middleware.GetUserInfoFromGin(c) // middleware sets authenticated user

	db := database.DB.WithContext(c.Request.Context())

	// Chỉ cho phép user xem của chính mình hoặc manager team của user đó
	if authUserID != userID && !isManagerOfSameTeam(c, authUserID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var ownedFolders []models.Folder
	db.Where("owner_id = ?", userID).Find(&ownedFolders)

	var sharedFolders []models.Folder
	db.Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", userID).Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"owned_folders":  ownedFolders,
		"shared_folders": sharedFolders,
	})
}

func isManagerOfTeam(c *gin.Context, userID, teamID string) bool {
	db := database.DB.WithContext(c.Request.Context())
	var roster models.Roster
	if err := db.Where("user_id = ? AND team_id = ? AND is_leader = true", userID, teamID).First(&roster).Error; err != nil {
		return false
	}
	return true
}

func isManagerOfSameTeam(c *gin.Context, managerID, memberID string) bool {
	db := database.DB.WithContext(c.Request.Context())
	var managerRosters []models.Roster
	db.Where("user_id = ? AND is_leader = true", managerID).Find(&managerRosters)

	var memberRosters []models.Roster
	db.Where("user_id = ?", memberID).Find(&memberRosters)

	for _, m := range managerRosters {
		for _, u := range memberRosters {
			if m.TeamID == u.TeamID {
				return true
			}
		}
	}
	return false
}
