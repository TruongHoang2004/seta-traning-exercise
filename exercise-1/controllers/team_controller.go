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

// @Security BearerAuth
// @Summary Create a new team
// @Description Create a new team with managers and members
// @Tags teams
// @Accept json
// @Produce json
// @Param team body dto.CreateTeamInput true "Team data with managers and members"
// @Success 201 {object} object "Created team with roster"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "User not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams [post]
func CreateTeam(c *gin.Context) {
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Lấy thông tin user từ DB
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can create teams"})
		return
	}

	// Parse request body
	var input dto.CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate managers exist and have correct role
	var managerIDs []string
	for _, manager := range input.Managers {
		var managerUser models.User
		if err := database.DB.First(&managerUser, "id = ?", manager.ManagerID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("manager with ID %s not found", manager.ManagerID)})
			return
		}

		if managerUser.Role != models.RoleManager {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s is not a manager", manager.ManagerID)})
			return
		}

		managerIDs = append(managerIDs, manager.ManagerID)
	}

	// Validate members exist and have correct role
	var memberIDs []string
	for _, member := range input.Members {
		var memberUser models.User
		if err := database.DB.First(&memberUser, "id = ?", member.MemberID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("member with ID %s not found", member.MemberID)})
			return
		}

		if memberUser.Role != models.RoleMember {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s is not a member", member.MemberID)})
			return
		}

		memberIDs = append(memberIDs, member.MemberID)
	}

	// Check for duplicate users between managers and members
	allUserIDs := append(managerIDs, memberIDs...)
	userMap := make(map[string]bool)
	for _, id := range allUserIDs {
		if userMap[id] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate user found in managers or members list"})
			return
		}
		userMap[id] = true
	}

	// Check if creator is already in the managers list
	creatorInList := false
	for _, id := range managerIDs {
		if id == userID {
			creatorInList = true
			break
		}
	}

	// Tạo team mới
	team := models.Team{
		TeamName: input.TeamName,
	}
	if err := database.DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	// Gán người tạo làm main manager (IsLeader = true) nếu chưa có trong danh sách
	if !creatorInList {
		roster := models.Roster{
			UserID:   userID,
			TeamID:   team.ID,
			IsLeader: true,
		}
		if err := database.DB.Create(&roster).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign creator as team leader"})
			return
		}
	}

	// Thêm managers vào roster
	for _, managerID := range managerIDs {
		isLeader := false
		// Nếu manager này là người tạo, set làm leader
		if managerID == userID {
			isLeader = true
		}

		roster := models.Roster{
			UserID:   managerID,
			TeamID:   team.ID,
			IsLeader: isLeader,
		}
		if err := database.DB.Create(&roster).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to add manager %s to team", managerID)})
			return
		}
	}

	// Thêm members vào roster
	for _, memberID := range memberIDs {
		roster := models.Roster{
			UserID:   memberID,
			TeamID:   team.ID,
			IsLeader: false,
		}
		if err := database.DB.Create(&roster).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to add member %s to team", memberID)})
			return
		}
	}

	// Load team with roster for response
	var teamWithRoster models.Team
	if err := database.DB.Preload("Rosters.User").First(&teamWithRoster, team.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load team data"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "team created successfully",
		"team":    teamWithRoster,
	})
}

// @Security BearerAuth
// @Summary Add member to team
// @Description Add a user as a member to a team
// @Tags teams
// @Accept json
// @Produce json
// @Param teamId path int true "Team ID"
// @Param input body dto.AddMemberToTeamInput true "Member data"
// @Success 201 {object} object "Success message"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams/{teamId}/members [post]
func AddMemberToTeam(c *gin.Context) {
	// 1. Get current user
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Get teamID from URL params
	teamID := c.Param("teamId")

	var currentUser models.User
	if err := database.DB.First(&currentUser, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "current user not found"})
		return
	}

	var input dto.AddMemberToTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Get team by teamID
	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	// 4. Get user by UserID
	var memberToAdd models.User
	if err := database.DB.First(&memberToAdd, "id = ?", input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to add not found"})
		return
	}

	// 5. Check if current user is in roster and is manager
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "current user is not in team roster"})
		return
	}

	if currentUser.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can add members to teams"})
		return
	}

	// 6. Only allow adding members (not managers)
	if memberToAdd.Role == models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "use the managers endpoint to add managers"})
		return
	}

	// 7. Check if user already exists in roster
	var existingRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", input.UserID, teamID).First(&existingRoster).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in the team"})
		return
	}

	// 8. Add to roster
	newMember := models.Roster{
		UserID:   input.UserID,
		TeamID:   teamID,
		IsLeader: false,
	}

	if err := database.DB.Create(&newMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member to team"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "member added successfully"})
}

// @Security BearerAuth
// @Summary Remove member from team
// @Description Remove a member from a team
// @Tags teams
// @Param teamId path int true "Team ID"
// @Param memberId path int true "Member ID"
// @Success 200 {object} object "Success message"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams/{teamId}/members/{memberId} [delete]
func RemoveMemberFromTeam(c *gin.Context) {
	// 1. Get current user
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Get teamID and memberID from URL params
	teamID := c.Param("teamId")
	memberID := c.Param("memberId")

	// 3. Check if current user is in the roster of the team
	var currentRoster models.Roster
	if err := database.DB.
		Where("user_id = ? AND team_id = ?", userID, teamID).
		First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this team"})
		return
	}

	// 4. Check if the target user is in the team
	var targetRoster models.Roster
	if err := database.DB.
		Where("user_id = ? AND team_id = ?", memberID, teamID).
		First(&targetRoster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not in the team"})
		return
	}

	// 5. Get target user info to check role
	var targetUser models.User
	if err := database.DB.First(&targetUser, "id = ?", memberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
		return
	}

	// 6. Only allow removing members (not managers)
	if targetUser.Role == models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "use the managers endpoint to remove managers"})
		return
	}

	// 7. Check permissions - only managers can remove members
	var currentUser models.User
	if err := database.DB.First(&currentUser, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "current user not found"})
		return
	}

	if currentUser.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can remove members"})
		return
	}

	// 8. Remove from roster
	if err := database.DB.Delete(&targetRoster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

// @Security BearerAuth
// @Summary Add manager to team
// @Description Add a user as a manager to a team
// @Tags teams
// @Accept json
// @Produce json
// @Param teamId path int true "Team ID"
// @Param input body dto.AddManagerToTeamInput true "Manager data"
// @Success 201 {object} object "Success message"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams/{teamId}/managers [post]
func AddManagerToTeam(c *gin.Context) {
	// 1. Get current user
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Get teamID from URL params
	teamID := c.Param("teamId")

	var input dto.AddManagerToTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Get team by teamID
	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	// 4. Get user by UserID
	var managerToAdd models.User
	if err := database.DB.First(&managerToAdd, "id = ?", input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to add not found"})
		return
	}

	// 5. Check if the user to add is actually a manager
	if managerToAdd.Role != models.RoleManager {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user must be a manager to be added as team manager"})
		return
	}

	// 6. Check if current user is team leader
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "current user is not in team roster"})
		return
	}

	if !currentRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "only team leader can add managers"})
		return
	}

	// 7. Check if user already exists in roster
	var existingRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", input.UserID, teamID).First(&existingRoster).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in the team"})
		return
	}

	// 8. Add to roster as manager (but not leader)
	newManager := models.Roster{
		UserID:   input.UserID,
		TeamID:   teamID,
		IsLeader: false,
	}

	if err := database.DB.Create(&newManager).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add manager to team"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "manager added successfully"})
}

// @Security BearerAuth
// @Summary Remove manager from team
// @Description Remove a manager from a team
// @Tags teams
// @Param teamId path int true "Team ID"
// @Param managerId path int true "Manager ID"
// @Success 200 {object} object "Success message"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams/{teamId}/managers/{managerId} [delete]
func RemoveManagerFromTeam(c *gin.Context) {
	// 1. Get current user
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Get teamID and managerID from URL params
	teamID := c.Param("teamId")
	managerID := c.Param("managerId")

	// 3. Check if current user is team leader
	var currentRoster models.Roster
	if err := database.DB.
		Where("user_id = ? AND team_id = ?", userID, teamID).
		First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this team"})
		return
	}

	if !currentRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "only team leader can remove managers"})
		return
	}

	// 4. Check if the target user is in the team
	var targetRoster models.Roster
	if err := database.DB.
		Where("user_id = ? AND team_id = ?", managerID, teamID).
		First(&targetRoster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not in the team"})
		return
	}

	// 5. Get target user info to verify they are a manager
	var targetUser models.User
	if err := database.DB.First(&targetUser, "id = ?", managerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
		return
	}

	if targetUser.Role != models.RoleManager {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target user is not a manager"})
		return
	}

	// 6. Prevent leader from removing themselves
	if userID == managerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "team leader cannot remove themselves"})
		return
	}

	// 7. Prevent removing the team leader
	if targetRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot remove team leader"})
		return
	}

	// 8. Remove from roster
	if err := database.DB.Delete(&targetRoster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove manager"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "manager removed successfully"})
}
