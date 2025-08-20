package persistence

import (
	"collab-service/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// Repository implementation
type RosterRepositoryImpl struct {
	db *gorm.DB
}

// Constructor
func NewRosterRepository(db *gorm.DB) entity.RosterRepository {
	return &RosterRepositoryImpl{db: db}
}

// Create implements entity.RosterRepository
func (r *RosterRepositoryImpl) Create(ctx context.Context, roster *entity.Roster) error {
	model := RosterModelFromDomain(roster)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID implements entity.RosterRepository
func (r *RosterRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Roster, error) {
	var model RosterModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// ListByTeamID implements entity.RosterRepository
func (r *RosterRepositoryImpl) ListByTeamID(ctx context.Context, teamID string) ([]*entity.Roster, error) {
	var models []RosterModel
	if err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&models).Error; err != nil {
		return nil, err
	}

	rosters := make([]*entity.Roster, len(models))
	for i, m := range models {
		rosters[i] = m.ToDomain()
	}
	return rosters, nil
}

// Delete implements entity.RosterRepository
func (r *RosterRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&RosterModel{}, "id = ?", id).Error
}
