package dto

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// CreateTeamRequest represents the request payload for creating a team
type CreateTeamRequest struct {
	TeamName string      `json:"teamName" binding:"required"`
	Managers []uuid.UUID `json:"managers,omitempty"`
	Members  []uuid.UUID `json:"members,omitempty"`
}

type AddMembersRequest struct {
	MemberIDs []uuid.UUID `json:"memberIds" binding:"required"`
}
type UpdateTeamRequest struct {
	Team entity.Team `json:"team" binding:"required"`
}

// RosterID represents a roster reference in team requests
type RosterID struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// TeamResponse represents the response payload for team operations
type TeamResponse struct {
	ID       uuid.UUID `json:"id"`
	TeamName string    `json:"teamName"`
	// Rosters   []RosterResponse `json:"rosters,omitempty"`
	Members   []uuid.UUID `json:"members,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

type RosterResponse struct {
	UserID uuid.UUID `json:"userId"`
	// Role   entity.TeamAccessRole `json:"role"`
}

// TeamListResponse represents a list of teams
type TeamListResponse struct {
	Teams []TeamResponse `json:"teams"`
}

func ToResponse(team *entity.Team) *TeamResponse {
	if team == nil {
		return nil
	}

	var memberIDs []uuid.UUID
	for _, roster := range team.Rosters {
		memberIDs = append(memberIDs, roster.UserID)
	}

	return &TeamResponse{
		ID:       team.ID,
		TeamName: team.Name,
		// Rosters:   rosterResponses,
		Members:   memberIDs,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}
}
