package handler

import (
	"collab-service/internal/application"
	"collab-service/internal/interface/http/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamService *application.TeamService
}

func NewTeamHandler(service *application.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: service,
	}
}

// @Security BearerAuth
// @Summary Create a new team
// @Description Create a new team with the given name and members
// @Tags teams
// @Accept json
// @Produce json
// @Param request body dto.CreateTeamRequest true "Team creation request"
// @Router /teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var request dto.CreateTeamRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	team, err := h.teamService.CreateTeam(c, request.TeamName, request.Managers, request.Members)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response dto.TeamResponse
	response.ID = team.ID
	response.TeamName = team.Name
	response.Rosters = team.Rosters

	c.JSON(http.StatusCreated, response)
}

// @Security BearerAuth
// @Summary Get team by ID
// @Description Get team details by ID
// @Tags teams
// @Param teamId path string true "Team ID (UUID)"
// @Success 200 {object} entity.Team "Team details"
// @Failure 400 {object} object "Bad request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 404 {object} object "Team not found"
// @Failure 500 {object} object "Internal server error"
// @Router /teams/{teamId} [get]
func (h *TeamHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Parse param thành UUID
	teamID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	team, err := h.teamService.GetTeamByID(c, teamID)
	if err != nil {
		application.HandleError(c, err)
		return
	}
	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var response dto.TeamResponse
	response.ID = team.ID
	response.TeamName = team.Name
	response.Rosters = team.Rosters

	c.JSON(http.StatusOK, response)
}

// @Security BearerAuth
// @Summary Get all teams of the current user
// @Description Get all teams that the current user is a member of
// @Tags teams
// @Success 200 {array} dto.TeamResponse "List of teams"
// @Failure 500 {object} object "Internal server error"
// @Router /teams [get]
func (h *TeamHandler) GetAllByUserID(c *gin.Context) {
	teams, err := h.teamService.GetAllTeamsOfUser(c)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response []dto.TeamResponse
	for _, team := range teams {
		response = append(response, dto.TeamResponse{
			ID:       team.ID,
			TeamName: team.Name,
			Rosters:  team.Rosters,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Security BearerAuth
// @Summary Add members to a team
// @Description Add new members to an existing team
// @Tags teams
// @Accept json
// @Produce json
// @Param request body dto.AddMembersRequest true "Add members request"
// @Router /teams/{teamId}/members [post]
func (h *TeamHandler) AddMembers(c *gin.Context) {
	var request dto.AddMembersRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	team, err := h.teamService.AddMembers(c, request.TeamID, request.MemberIDs)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response dto.TeamResponse
	response.ID = team.ID
	response.TeamName = team.Name
	response.Rosters = team.Rosters

	c.JSON(http.StatusOK, response)
}

func (h *TeamHandler) AddManager(c *gin.Context) {
	var request dto.AddManagerRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	team, err := h.teamService.AddManager(c, request.TeamID, request.ManagerID)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response dto.TeamResponse
	response.ID = team.ID
	response.TeamName = team.Name
	response.Rosters = team.Rosters

	c.JSON(http.StatusOK, response)
}

// @Security BearerAuth
// @Summary Remove a member from a team
// @Description Remove a member from an existing team
// @Tags teams
// @Accept json
// @Produce json
// @Param request body dto.RemoveMemberRequest true "Remove member request"
// @Router /teams/{teamId}/members/{memberId} [delete]
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	var request dto.RemoveMemberRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.teamService.RemoveMember(c, request.TeamID, request.MemberID); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TeamHandler) RemoveManager(c *gin.Context) {
	var request dto.RemoveManagerRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.teamService.RemoveManager(c, request.TeamID, request.ManagerID); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// @Security BearerAuth
// @Summary Update a team
// @Description Update an existing team
// @Tags teams
// @Accept json
// @Produce json
// @Param request body dto.UpdateTeamRequest true "Update team request"
// @Router /teams/{teamId} [put]
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	var request dto.UpdateTeamRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	team, err := h.teamService.UpdateTeam(c, &request.Team)
	if err != nil {
		application.HandleError(c, err)
		return
	}

	var response dto.TeamResponse
	response.ID = team.ID
	response.TeamName = team.Name
	response.Rosters = team.Rosters

	c.JSON(http.StatusOK, response)
}

// @Security BearerAuth
// @Summary Delete a team
// @Description Delete an existing team
// @Tags teams
// @Router /teams/{teamId} [delete]
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	id := c.Param("id")

	// Parse param thành UUID
	teamID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	if err := h.teamService.DeleteTeam(c, teamID); err != nil {
		application.HandleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
