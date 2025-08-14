package controllers

import (
	"collab-service/config"
	"collab-service/internal/database"
	"collab-service/internal/dto"
	"collab-service/internal/middleware"
	"collab-service/internal/models"
	"collab-service/pkg/cache"
	"collab-service/pkg/client"
	"collab-service/pkg/logger"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	// 1. Xác thực người dùng và quyền hạn
	userID, userRole := middleware.GetUserInfoFromGin(c)

	if userRole != client.UserType(client.UserTypeManager) {
		logger.Info("Unauthorized access", "userId", userID, "role", userRole)
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can create teams"})
		return
	}

	// 2. Đọc input
	var input dto.CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Info("Invalid input", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Thu thập danh sách ID của manager và member
	var managerIDs, memberIDs []string
	for _, m := range input.Managers {
		managerIDs = append(managerIDs, m.ManagerID)
	}
	for _, m := range input.Members {
		memberIDs = append(memberIDs, m.MemberID)
	}

	// 4. Khởi tạo GraphQL client
	userClient := client.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)

	// 5. Kiểm tra và lấy thông tin các manager
	if len(managerIDs) > 0 {
		for _, managerID := range managerIDs {
			manager, err := userClient.GetUser(managerID)
			if err != nil {
				logger.Info("Manager not found", "managerId", managerID, "error", err)
				c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("manager %s not found", managerID)})
				return
			}
			if manager.Role != client.UserTypeManager {
				logger.Info("Invalid user role", "userId", manager.UserID, "role", manager.Role)
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s is not a manager", manager.UserID)})
				return
			}
		}
	}

	// 6. Kiểm tra và lấy thông tin các member
	if len(memberIDs) > 0 {
		for _, memberID := range memberIDs {
			member, err := userClient.GetUser(memberID)
			if err != nil {
				logger.Error("Member not found", "memberId", memberID, "error", err)
				c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("member %s not found", memberID)})
				return
			}
			if member.Role != client.UserTypeMember {
				logger.Info("Invalid user role", "userId", member.UserID, "role", member.Role)
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s is not a member", member.UserID)})
				return
			}
		}
	}

	// 7. Kiểm tra trùng lặp ID giữa managers và members
	userMap := make(map[string]bool)
	for _, id := range append(managerIDs, memberIDs...) {
		if userMap[id] {
			logger.Error("Duplicate user in managers or members list", "userId", id)
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate user in managers or members list"})
			return
		}
		userMap[id] = true
	}

	// 8. Kiểm tra xem người tạo đã có trong managers chưa
	creatorInList := false
	for _, id := range managerIDs {
		if id == userID {
			creatorInList = true
			break
		}
	}

	// 9. Tạo transaction và thêm team mới
	tx := database.DB.Begin()
	team := models.Team{TeamName: input.TeamName}
	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create team", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	// 10. Chuẩn bị danh sách các roster mới
	var rosters []models.Roster
	// Thêm creator làm leader nếu chưa có
	if !creatorInList {
		rosters = append(rosters, models.Roster{UserID: userID, TeamID: team.ID, IsLeader: true})
	}
	// Thêm managers (đánh dấu leader nếu trùng với creator)
	for _, id := range managerIDs {
		isLeader := (id == userID)
		rosters = append(rosters, models.Roster{UserID: id, TeamID: team.ID, IsLeader: isLeader})
	}
	// Thêm members (không phải leader)
	for _, id := range memberIDs {
		rosters = append(rosters, models.Roster{UserID: id, TeamID: team.ID, IsLeader: false})
	}

	// 11. Thêm tất cả các roster trong một lần insert
	if len(rosters) > 0 {
		if err := tx.Create(&rosters).Error; err != nil {
			tx.Rollback()
			logger.Error("Failed to create team rosters", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign members to team"})
			return
		}
	}

	// 12. Commit transaction
	tx.Commit()

	// 13. Trả về dữ liệu team kèm danh sách roster (không còn eager load User)
	var teamWithRoster models.Team
	if err := database.DB.Preload("Rosters").
		First(&teamWithRoster, "id = ?", team.ID).Error; err != nil {
		logger.Error("Failed to load team with roster", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load team data"})
		return
	}

	logger.Info("Team created successfully", "teamId", team.ID, "teamName", team.TeamName)
	c.JSON(http.StatusCreated, gin.H{"message": "team created successfully", "team": teamWithRoster})
}

// @Security BearerAuth
// @Summary Get team by ID
// @Description Get team details by ID
// @Tags teams
// @Param teamId path string true "Team ID (UUID)"
// @Success 200 {object} models.Team "Team details"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 404 {object} object "Team not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams/{teamId} [get]
func GetTeamByID(c *gin.Context) {
	teamID := c.Param("teamId")

	// 1. Thử lấy userIDs từ cache
	userIDs, err := cache.GetTeamMembers(c, teamID)
	if err == nil && len(userIDs) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"teamId":  teamID,
			"members": userIDs,
			"source":  "cache",
		})
		return
	}

	// 2. Nếu không có trong cache → query DB
	var team models.Team
	if err := database.DB.Preload("Rosters").First(&team, "id = ?", teamID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("Not found")
		} else {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("Found team:", team)
	}

	c.JSON(http.StatusOK, gin.H{
		"teamId": teamID,
		"members": func() []string {
			ids := []string{}
			for _, r := range team.Rosters {
				ids = append(ids, r.UserID)
			}
			return ids
		}(),
		"source": "db",
	})

	// 3. Cache lại userIDs
	for _, r := range team.Rosters {
		_ = cache.AddTeamMember(c, teamID, r.UserID)
	}

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
	// 1. Xác thực người dùng
	userID, _ := middleware.GetUserInfoFromGin(c)

	// 2. Lấy teamID từ URL
	teamID := c.Param("teamId")

	// 3. Lấy input
	var input dto.AddMemberToTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Info("Invalid input", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Kiểm tra team tồn tại
	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		logger.Info("Team not found", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	// 5. Khởi tạo GraphQL client
	userClient := client.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)

	// 6. Lấy thông tin người muốn thêm
	memberToAdd, err := userClient.GetUser(input.UserID)
	if err != nil {
		logger.Info("User to add not found", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "user to add not found"})
		return
	}

	// 7. Kiểm tra user hiện tại có trong roster
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).
		First(&currentRoster).Error; err != nil {
		logger.Info("Current user not in team roster", "error", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "current user is not in team roster"})
		return
	}

	// 8. Lấy thông tin user hiện tại và kiểm tra quyền
	currentUser, err := userClient.GetUser(userID)
	if err != nil {
		logger.Error("Failed to get current user details", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current user details"})
		return
	}

	if currentUser.Role != client.UserTypeManager {
		logger.Info("Unauthorized access", "userId", userID, "role", currentUser.Role)
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can add members to teams"})
		return
	}

	// 9. Chỉ được thêm member (không phải manager)
	if memberToAdd.Role == client.UserTypeManager {
		logger.Info("Cannot add manager as member", "userId", memberToAdd.UserID)
		c.JSON(http.StatusForbidden, gin.H{"error": "use the managers endpoint to add managers"})
		return
	}

	// 10. Kiểm tra nếu người dùng đã tồn tại trong roster
	var existingRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", input.UserID, teamID).
		First(&existingRoster).Error; err == nil {
		logger.Info("User already in team roster", "userId", input.UserID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in the team"})
		return
	}

	// 11. Thêm thành viên mới vào roster
	newMember := models.Roster{UserID: input.UserID, TeamID: teamID, IsLeader: false}
	if err := database.DB.Create(&newMember).Error; err != nil {
		logger.Error("Failed to add member to team", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member to team"})
		return
	}

	logger.Info("Member added to team", "teamId", teamID, "userId", input.UserID)
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
	// 1. Xác thực người dùng
	userID, userRole := middleware.GetUserInfoFromGin(c)

	// 2. Lấy teamID và memberID từ URL
	teamID := c.Param("teamId")
	memberID := c.Param("memberId")

	// 3. Kiểm tra user hiện tại có trong đội
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).
		First(&currentRoster).Error; err != nil {
		logger.Info("User not in team roster", "userId", userID, "teamId", teamID, "error", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this team"})
		return
	}

	// 4. Kiểm tra quyền của user hiện tại
	if client.UserType(userRole) != client.UserTypeManager {
		logger.Info("Unauthorized access", "userId", userID, "role", userRole)
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can remove members"})
		return
	}

	// 5. Kiểm tra user mục tiêu có trong đội
	var targetRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", memberID, teamID).
		First(&targetRoster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not in the team"})
		return
	}

	// 6. Xóa khỏi roster
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
	// 1. Xác thực user
	userID, _ := middleware.GetUserInfoFromGin(c)

	// 2. Lấy teamID
	teamID := c.Param("teamId")

	// 3. Đọc input
	var input dto.AddManagerToTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Kiểm tra team tồn tại
	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	// 5. Khởi tạo GraphQL client
	userClient := client.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)

	// 6. Kiểm tra user cần thêm tồn tại và là manager
	managerToAdd, err := userClient.GetUser(input.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to add not found"})
		return
	}
	if managerToAdd.Role != client.UserTypeManager {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user must be a manager to be added as team manager"})
		return
	}

	// 7. Kiểm tra user hiện tại là leader của team
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "current user is not in team roster"})
		return
	}
	if !currentRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "only team leader can add managers"})
		return
	}

	// 8. Kiểm tra user đã tồn tại trong team chưa
	var existingRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", input.UserID, teamID).First(&existingRoster).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in the team"})
		return
	}

	// 9. Thêm manager mới (không phải leader)
	newManager := models.Roster{UserID: input.UserID, TeamID: teamID, IsLeader: false}
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
	// 1. Xác thực người dùng
	userID, _ := middleware.GetUserInfoFromGin(c)

	// 2. Lấy teamID và managerID từ URL
	teamID := c.Param("teamId")
	managerID := c.Param("managerId")

	// 3. Kiểm tra user hiện tại là leader của team
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this team"})
		return
	}
	if !currentRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "only team leader can remove managers"})
		return
	}

	// 4. Kiểm tra manager mục tiêu có trong team
	var targetRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", managerID, teamID).
		First(&targetRoster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not in the team"})
		return
	}

	// 5. Kiểm tra target không phải leader
	if targetRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot remove team leader"})
		return
	}

	// 6. Xóa khỏi roster
	if err := database.DB.Delete(&targetRoster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove manager"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "manager removed successfully"})
}
