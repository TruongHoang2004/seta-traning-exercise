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
	TeamID    uuid.UUID   `json:"teamId" binding:"required"`
	MemberIDs []uuid.UUID `json:"memberIds" binding:"required"`
}

type RemoveMemberRequest struct {
	TeamID   uuid.UUID `json:"teamId" binding:"required"`
	MemberID uuid.UUID `json:"memberId" binding:"required"`
}

type AddManagerRequest struct {
	TeamID    uuid.UUID `json:"teamId" binding:"required"`
	ManagerID uuid.UUID `json:"managerId" binding:"required"`
}

type RemoveManagerRequest struct {
	TeamID    uuid.UUID `json:"teamId" binding:"required"`
	ManagerID uuid.UUID `json:"managerId" binding:"required"`
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
	ID        uuid.UUID       `json:"id"`
	TeamName  string          `json:"teamName"`
	Rosters   []entity.Roster `json:"rosters,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// TeamListResponse represents a list of teams
type TeamListResponse struct {
	Teams []TeamResponse `json:"teams"`
}
