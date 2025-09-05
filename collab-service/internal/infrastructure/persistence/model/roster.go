package model

import (
	"collab-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type RosterModel struct {
	ID        uuid.UUID             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    uuid.UUID             `gorm:"type:uuid;index;uniqueIndex:idx_team_user"`
	TeamID    uuid.UUID             `gorm:"type:uuid;index;uniqueIndex:idx_team_user"`
	Team      *TeamModel            `gorm:"foreignKey:TeamID;references:ID"`
	Role      entity.TeamAccessRole `gorm:"type:varchar(20);default:'MEMBER'"`
	CreatedAt time.Time             `gorm:"autoCreateTime"`
	UpdatedAt time.Time             `gorm:"autoUpdateTime"`

	// Unique constraint để không cho phép duplicate (team_id, user_id)
	_ struct{} `gorm:"uniqueIndex:idx_team_user,composite:team_id,user_id"`
}

func (RosterModel) TableName() string {
	return "rosters"
}

// Convert RosterModel -> domain.Roster
func (m *RosterModel) ToDomain() *entity.Roster {
	return &entity.Roster{
		ID:     m.ID,
		UserID: m.UserID,
		TeamID: m.TeamID,
		Role:   m.Role,
	}
}

// Convert entity.Roster -> RosterModel
func RosterModelFromDomain(r *entity.Roster) *RosterModel {
	return &RosterModel{
		ID:     r.ID,
		UserID: r.UserID,
		TeamID: r.TeamID,
		Role:   r.Role,
	}
}
