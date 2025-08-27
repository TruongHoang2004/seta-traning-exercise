package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TeamAccessRole string

const (
	TeamOwner   TeamAccessRole = "OWNER"
	TeamManager TeamAccessRole = "MANAGER"
	TeamMember  TeamAccessRole = "MEMBER"
	TeamNone    TeamAccessRole = "NONE"
)

// IsHigherOrEqualTo checks if the current role has equal or higher privileges than the given role
func (r TeamAccessRole) IsHigherOrEqualTo(other TeamAccessRole) bool {
	roleWeight := map[TeamAccessRole]int{
		TeamOwner:   3,
		TeamManager: 2,
		TeamMember:  1,
	}

	return roleWeight[r] >= roleWeight[other]
}

// Roster represents a user's membership in a team
type Roster struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TeamID    uuid.UUID
	Team      *Team
	Role      TeamAccessRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewRoster creates a new roster entry
func NewRoster(userID uuid.UUID, team *Team, role TeamAccessRole) *Roster {
	return &Roster{
		UserID:    userID,
		Team:      team,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type RosterRepository interface {
	// Save a new roster
	Create(ctx context.Context, roster *Roster) error

	// Get a roster by ID
	GetByID(ctx context.Context, id string) (*Roster, error)

	// List rosters by team ID
	ListByTeamID(ctx context.Context, teamID string) ([]*Roster, error)

	// Delete a roster
	Delete(ctx context.Context, id string) error
}
