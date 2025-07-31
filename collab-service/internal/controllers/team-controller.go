package controllers

import (
	"collab-service/internal/database"
	"collab-service/internal/dto"
	"collab-service/internal/middleware"
	"collab-service/internal/models"
	"collab-service/pkg/client"
	"fmt"
	"net/http"

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
	// 1. Xác thực người dùng và quyền hạn
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var creator models.User
	if err := database.DB.First(&creator, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if creator.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can create teams"})
		return
	}

	// 2. Đọc input
	var input dto.CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
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

	// 4. Kiểm tra và lấy thông tin các manager theo một truy vấn duy nhất
	var managers []models.User
	if len(managerIDs) > 0 {
		database.DB.Where("id IN ?", managerIDs).Find(&managers)
		if len(managers) != len(managerIDs) {
			c.JSON(http.StatusNotFound, gin.H{"error": "some manager IDs not found"})
			return
		}
		for _, u := range managers {
			if u.Role != models.RoleManager {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s is not a manager", u.ID)})
				return
			}
		}
	}

	// 5. Kiểm tra và lấy thông tin các member theo một truy vấn duy nhất
	var members []models.User
	if len(memberIDs) > 0 {
		database.DB.Where("id IN ?", memberIDs).Find(&members)
		if len(members) != len(memberIDs) {
			c.JSON(http.StatusNotFound, gin.H{"error": "some member IDs not found"})
			return
		}
		for _, u := range members {
			if u.Role != models.RoleMember {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s is not a member", u.ID)})
				return
			}
		}
	}

	// 6. Kiểm tra trùng lặp ID giữa managers và members
	userMap := make(map[string]bool)
	for _, id := range append(managerIDs, memberIDs...) {
		if userMap[id] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate user in managers or members list"})
			return
		}
		userMap[id] = true
	}

	// 7. Kiểm tra xem người tạo đã có trong managers chưa
	creatorInList := false
	for _, id := range managerIDs {
		if id == userID {
			creatorInList = true
			break
		}
	}

	// 8. Tạo transaction và thêm team mới
	tx := database.DB.Begin()
	team := models.Team{TeamName: input.TeamName}
	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	// 9. Chuẩn bị danh sách các roster mới
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

	// 10. Thêm tất cả các roster trong một lần insert
	if len(rosters) > 0 {
		if err := tx.Create(&rosters).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign members to team"})
			return
		}
	}

	// 11. Commit transaction
	tx.Commit()

	// 12. Trả về dữ liệu team kèm danh sách roster (eager load user)
	var teamWithRoster models.Team
	if err := database.DB.Preload("Rosters.User").First(&teamWithRoster, team.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load team data"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "team created successfully", "team": teamWithRoster})
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
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Lấy teamID từ URL
	teamID := c.Param("teamId")

	// 3. Lấy input
	var input dto.AddMemberToTeamInput
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

	// 5. Lấy thông tin người muốn thêm
	userClient := client.NewGraphQLClient("http://localhost:4000/query")
	memberToAdd, err := userClient.GetUser(input.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to add not found"})
		return
	}

	// Convert client.User to models.User
	// 6. Kiểm tra user hiện tại có trong roster và là manager
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).
		First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "current user is not in team roster"})
		return
	}

	// Get current user details
	currentUser, err := userClient.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current user details"})
		return
	}

	if currentUser.Role != client.UserTypeManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "only managers can add members to teams"})
		return
	}

	// 7. Chỉ được thêm member (không phải manager)
	if memberToAdd.Role == client.UserTypeManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "use the managers endpoint to add managers"})
		return
	}

	// 8. Kiểm tra nếu người dùng đã tồn tại trong roster
	var existingRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", input.UserID, teamID).
		First(&existingRoster).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in the team"})
		return
	}

	// 9. Thêm thành viên mới vào roster
	newMember := models.Roster{UserID: input.UserID, TeamID: teamID, IsLeader: false}
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
	// 1. Xác thực người dùng
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Lấy teamID và memberID từ URL
	teamID := c.Param("teamId")
	memberID := c.Param("memberId")

	// 3. Kiểm tra user hiện tại có trong đội (và là manager) (eager load User)
	var currentRoster models.Roster
	if err := database.DB.Preload("User").
		Where("user_id = ? AND team_id = ?", userID, teamID).
		First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this team"})
		return
	}
	// if currentRoster.User.Role != models.RoleManager {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "only managers can remove members"})
	// 	return
	// }

	// 4. Kiểm tra user mục tiêu có trong đội (và lấy thông tin role) (eager load User)
	var targetRoster models.Roster
	if err := database.DB.Preload("User").
		Where("user_id = ? AND team_id = ?", memberID, teamID).
		First(&targetRoster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not in the team"})
		return
	}
	// if targetRoster.User.Role == models.RoleManager {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "use the managers endpoint to remove managers"})
	// 	return
	// }

	// 5. Xóa khỏi roster
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
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

	// 5. Kiểm tra user cần thêm tồn tại và là manager
	var managerToAdd models.User
	if err := database.DB.First(&managerToAdd, "id = ?", input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to add not found"})
		return
	}
	if managerToAdd.Role != models.RoleManager {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user must be a manager to be added as team manager"})
		return
	}

	// 6. Kiểm tra user hiện tại là leader của team
	var currentRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).First(&currentRoster).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "current user is not in team roster"})
		return
	}
	if !currentRoster.IsLeader {
		c.JSON(http.StatusForbidden, gin.H{"error": "only team leader can add managers"})
		return
	}

	// 7. Kiểm tra user đã tồn tại trong team chưa
	var existingRoster models.Roster
	if err := database.DB.Where("user_id = ? AND team_id = ?", input.UserID, teamID).First(&existingRoster).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in the team"})
		return
	}

	// 8. Thêm manager mới (không phải leader)
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
	userID, err := middleware.GetUserIDFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

	// 4. Kiểm tra manager mục tiêu có trong team (và lấy thông tin role, isLeader)
	var targetRoster models.Roster
	if err := database.DB.Preload("User").
		Where("user_id = ? AND team_id = ?", managerID, teamID).
		First(&targetRoster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not in the team"})
		return
	}
	// 5. Kiểm tra target là manager và không phải leader
	// if targetRoster.User.Role != models.RoleManager {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "target user is not a manager"})
	// 	return
	// }
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
