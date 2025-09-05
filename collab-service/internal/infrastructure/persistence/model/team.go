package model

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type TeamModel struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TeamName string
	Rosters  []RosterModel `gorm:"foreignKey:TeamID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type CachedMember struct {
	UserID uuid.UUID             `json:"user_id"`
	Role   entity.TeamAccessRole `json:"role"`
}

type CachedTeam struct {
	ID      uuid.UUID      `json:"id"`
	Name    string         `json:"name"`
	Members []CachedMember `json:"members"`
}

func (TeamModel) TableName() string {
	return "teams"
}

// Convert Team -> domain.Team
func (m *TeamModel) ToDomain() *entity.Team {
	rosters := make([]entity.Roster, len(m.Rosters))
	for i, r := range m.Rosters {
		rosters[i] = *r.ToDomain()
	}

	return &entity.Team{
		ID:        m.ID,
		Name:      m.TeamName,
		Rosters:   rosters,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// Convert entity.Team -> Team
func TeamModelFromDomain(t *entity.Team) *TeamModel {
	rosters := make([]RosterModel, len(t.Rosters))
	for i, r := range t.Rosters {
		rosters[i] = *RosterModelFromDomain(&r)
	}
	return &TeamModel{
		TeamName:  t.Name,
		Rosters:   rosters,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
