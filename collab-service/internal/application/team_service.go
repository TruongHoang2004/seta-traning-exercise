package application

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/event"
	"collab-service/internal/infrastructure/logger"
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
	eventProducer  *event.TeamActivityProducer
}

// NewTeamService creates a new team service instance
func NewTeamService(repo entity.TeamRepository, userRepo entity.UserRepository) *TeamService {
	return &TeamService{
		teamRepository: repo,
		userRepository: userRepo,
		eventProducer:  event.GetTeamActivityProducer(),
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
	savedTeam, err := s.teamRepository.Create(c.Request.Context(), team)

	go func() {
		s.eventProducer.Produce(event.NewTeamEvent(event.TeamCreated, savedTeam.ID.String(), userID.String(), ""))
		for _, memberID := range members {
			s.eventProducer.Produce(event.NewTeamEvent(event.MemberAdded, savedTeam.ID.String(), userID.String(), memberID.String()))
		}
		for _, managerID := range manager {
			s.eventProducer.Produce(event.NewTeamEvent(event.ManagerAdded, savedTeam.ID.String(), userID.String(), managerID.String()))
		}
	}()

	return savedTeam, err
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
	userID, _ := middleware.GetUserInfoFromGin(c)
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

	go func() {
		for _, member := range members {
			if err := s.eventProducer.Produce(event.NewTeamEvent(event.MemberAdded, team.ID.String(), userID.String(), member.String())); err != nil {
				logger.Error("failed to produce event: %v", err)
			}
		}
	}()

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

	if !role.IsHigherOrEqualTo(entity.TeamManager) {
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

	go s.eventProducer.Produce(event.NewTeamEvent(event.ManagerAdded, team.ID.String(), userId.String(), managerID.String()))
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

	go s.eventProducer.Produce(event.NewTeamEvent(event.MemberRemoved, teamID.String(), userId.String(), memberID.String()))
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

	targetRole, err := s.teamRepository.GetRole(c.Request.Context(), teamID, managerID)
	switch targetRole {
	case entity.TeamOwner:
		return NewForbiddenError("cannot remove team owner")
	case entity.TeamNone:
		return NewNotFoundError(fmt.Sprintf("user %s is not a member of team %s", managerID, teamID))
	}

	if err != nil {
		return err
	}

	if err := s.teamRepository.RemoveMember(c.Request.Context(), teamID, managerID); err != nil {
		return err
	}

	go s.eventProducer.Produce(event.NewTeamEvent(event.ManagerRemoved, teamID.String(), userId.String(), managerID.String()))

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
