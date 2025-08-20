package application

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/interface/http/middleware"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TeamService handles team-related operations
type TeamService struct {
	teamRepository entity.TeamRepository
	userRepository entity.UserRepository
}

// NewTeamService creates a new team service instance
func NewTeamService(repo entity.TeamRepository, userRepo entity.UserRepository) *TeamService {
	return &TeamService{
		teamRepository: repo,
		userRepository: userRepo,
	}
}

// CreateTeam creates a new team
func (s *TeamService) CreateTeam(c *gin.Context, name string, members []uuid.UUID, manager []uuid.UUID) (*entity.Team, error) {
	// Get the user ID from the context
	userID, _ := middleware.GetUserInfoFromGin(c)

	// Create a new team with the current user as creator
	team := entity.NewTeam(name, []entity.Roster{})

	// Create a map to track all processed users to avoid duplicates
	processedUsers := make(map[uuid.UUID]bool)

	// **ADD CREATOR AS OWNER FIRST**
	creatorRoster := entity.NewRoster(userID, team, entity.TeamOwner)
	team.AddRoster(*creatorRoster)
	processedUsers[userID] = true

	// First process managers (they have higher priority)
	for _, managerID := range manager {
		if processedUsers[managerID] {
			continue // Skip if already processed
		}
		user, err := s.userRepository.GetByID(c.Request.Context(), managerID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("manager with ID %s not found", managerID)
		}
		roster := entity.NewRoster(managerID, team, entity.TeamManager)
		team.AddRoster(*roster)
		processedUsers[managerID] = true
	}

	// Then process regular members, skipping those who are already managers
	for _, memberID := range members {
		if processedUsers[memberID] {
			continue // Skip if already processed
		}
		user, err := s.userRepository.GetByID(c.Request.Context(), memberID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("member with ID %s not found", memberID)
		}
		roster := entity.NewRoster(memberID, team, entity.TeamMember)
		team.AddRoster(*roster)
		processedUsers[memberID] = true
	}

	// Save team and rosters
	return s.teamRepository.Create(c.Request.Context(), team)
}

func (s *TeamService) GetAllTeamsOfUser(c *gin.Context) ([]*entity.Team, error) {
	userID, _ := middleware.GetUserInfoFromGin(c)
	teams, err := s.teamRepository.GetAllByUserID(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *TeamService) GetTeamByID(c *gin.Context, id uuid.UUID) (*entity.Team, error) {
	team, err := s.teamRepository.GetByID(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	return team, nil
}

func (s *TeamService) AddMembers(c *gin.Context, teamID uuid.UUID, members []uuid.UUID) (*entity.Team, error) {
	team, err := s.teamRepository.GetByID(c.Request.Context(), teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, fmt.Errorf("team %s not found", teamID)
	}

	users, err := s.userRepository.List(c.Request.Context(), nil, members)
	if err != nil {
		return nil, err
	}

	if len(users) != len(members) {
		return nil, fmt.Errorf("some users not found")
	}

	var userIDs []uuid.UUID
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	err = s.teamRepository.AddMembers(c.Request.Context(), team.ID, userIDs)
	if err != nil {
		return nil, err
	}

	team, err = s.teamRepository.GetByID(c.Request.Context(), team.ID)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *TeamService) AddManager(c *gin.Context, teamID uuid.UUID, managerID uuid.UUID) (*entity.Team, error) {
	userId, _ := middleware.GetUserInfoFromGin(c)
	role, err := s.teamRepository.GetRole(c.Request.Context(), teamID, userId)
	if err != nil {
		return nil, err
	}

	if role != entity.TeamManager {
		return nil, NewForbiddenError("only team managers can add other managers")
	}

	team, err := s.teamRepository.GetByID(c.Request.Context(), teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, NewNotFoundError(fmt.Sprintf("team %s not found", teamID))
	}

	user, err := s.userRepository.GetByID(c.Request.Context(), managerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, NewNotFoundError(fmt.Sprintf("user %s not found", managerID))
	}

	if err := s.teamRepository.AddManager(c.Request.Context(), teamID, managerID); err != nil {
		return nil, err
	}

	team, err = s.teamRepository.GetByID(c.Request.Context(), teamID)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *TeamService) RemoveMember(c *gin.Context, teamID uuid.UUID, memberID uuid.UUID) error {
	userId, _ := middleware.GetUserInfoFromGin(c)
	role, err := s.teamRepository.GetRole(c.Request.Context(), teamID, userId)
	if err != nil {
		return err
	}

	if role == entity.TeamMember {
		// If the user is a regular member, they can only remove themselves
		if userId != memberID {
			return NewForbiddenError("regular members can only remove themselves")
		}
	}

	team, err := s.teamRepository.GetByID(c.Request.Context(), teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return NewNotFoundError(fmt.Sprintf("team %s not found", teamID))
	}

	if err := s.teamRepository.RemoveMember(c.Request.Context(), teamID, memberID); err != nil {
		return err
	}

	return nil
}

func (s *TeamService) RemoveManager(c *gin.Context, teamID uuid.UUID, managerID uuid.UUID) error {
	userId, _ := middleware.GetUserInfoFromGin(c)
	role, err := s.teamRepository.GetRole(c.Request.Context(), teamID, userId)
	if err != nil {
		return err
	}

	if role != entity.TeamOwner {
		return NewForbiddenError("only team managers can remove other managers")
	}

	team, err := s.teamRepository.GetByID(c.Request.Context(), teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return NewNotFoundError(fmt.Sprintf("team %s not found", teamID))
	}

	if err := s.teamRepository.RemoveMember(c.Request.Context(), teamID, managerID); err != nil {
		return err
	}

	return nil
}

func (s *TeamService) UpdateTeam(c *gin.Context, team *entity.Team) (*entity.Team, error) {

	team, err := s.teamRepository.GetByID(c.Request.Context(), team.ID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	// Update the team in the repository
	updatedTeam, err := s.teamRepository.Update(c.Request.Context(), team)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Team updated successfully: %v", updatedTeam)

	return updatedTeam, nil
}

func (s *TeamService) DeleteTeam(c *gin.Context, id uuid.UUID) error {

	userId, _ := middleware.GetUserInfoFromGin(c)
	role, err := s.teamRepository.GetRole(c.Request.Context(), id, userId)
	if err != nil {
		return err
	}

	if role != entity.TeamOwner {
		return NewForbiddenError("only team owners can delete a team")
	}

	team, err := s.teamRepository.GetByID(c.Request.Context(), id)
	if err != nil {
		return err
	}

	if team == nil {
		return NewNotFoundError(fmt.Sprintf("team %s not found", id))
	}

	if err := s.teamRepository.Delete(c.Request.Context(), id); err != nil {
		return err
	}

	return nil
}
