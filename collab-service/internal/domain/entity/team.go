package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Team represents a team entity in the domain
type Team struct {
	ID        uuid.UUID
	Name      string
	Rosters   []Roster
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTeam creates a new Team instance
func NewTeam(name string, rosters []Roster) *Team {
	return &Team{
		Name:      name,
		Rosters:   rosters,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (t *Team) AddRoster(roster Roster) {
	t.Rosters = append(t.Rosters, roster)
	roster.Team = t
}

type TeamRepository interface {
	Create(ctx context.Context, team *Team) (*Team, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetRole(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (TeamAccessRole, error)
	List(ctx context.Context) ([]*Team, error)
	AddMembers(ctx context.Context, teamID uuid.UUID, members []uuid.UUID) error
	AddManager(ctx context.Context, teamID uuid.UUID, managerID uuid.UUID) error
	RemoveMember(ctx context.Context, teamID uuid.UUID, memberID uuid.UUID) error
	Update(ctx context.Context, team *Team) (*Team, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
