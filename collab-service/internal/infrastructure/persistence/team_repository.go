package persistence

import (
	"collab-service/internal/domain/entity"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// Repository implementation
type TeamRepositoryImpl struct {
	db *gorm.DB
}

// Constructor
func NewTeamRepository(db *gorm.DB) entity.TeamRepository {
	return &TeamRepositoryImpl{db: db}
}

// Create implements entity.TeamRepository
func (r *TeamRepositoryImpl) Create(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	tx := r.db.WithContext(ctx).Begin()

	model := TeamModelFromDomain(team)

	// Đảm bảo roster IDs = nil để GORM generate
	for i := range model.Rosters {
		model.Rosters[i].ID = uuid.Nil
	}

	// GORM sẽ tự động create team và rosters
	if err := tx.Create(&model).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

// GetAllByUserID implements entity.TeamRepository
func (r *TeamRepositoryImpl) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Team, error) {
	var models []TeamModel
	if err := r.db.WithContext(ctx).
		Joins("JOIN rosters ON rosters.team_id = teams.id").
		Where("rosters.user_id = ?", userID).
		Preload("Rosters").
		Find(&models).Error; err != nil {
		return nil, err
	}

	teams := make([]*entity.Team, len(models))
	for i, m := range models {
		teams[i] = m.ToDomain()
	}
	return teams, nil
}

func (r *TeamRepositoryImpl) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&TeamModel{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByID implements entity.TeamRepository
func (r *TeamRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	var model TeamModel
	if err := r.db.WithContext(ctx).Preload("Rosters").First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// List implements entity.TeamRepository
func (r *TeamRepositoryImpl) List(ctx context.Context) ([]*entity.Team, error) {
	var models []TeamModel
	if err := r.db.WithContext(ctx).Preload("Rosters").Find(&models).Error; err != nil {
		return nil, err
	}

	teams := make([]*entity.Team, len(models))
	for i, m := range models {
		teams[i] = m.ToDomain()
	}
	return teams, nil
}

// Get role
func (r *TeamRepositoryImpl) GetRole(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (entity.TeamAccessRole, error) {
	var roster struct {
		Role string
	}

	err := r.db.WithContext(ctx).
		Table("rosters").
		Select("role").
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&roster).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.TeamNone, nil
		}
		return "", err
	}

	return entity.TeamAccessRole(roster.Role), nil
}

// Add implements entity.TeamRepository
func (r *TeamRepositoryImpl) AddMembers(ctx context.Context, teamID uuid.UUID, members []uuid.UUID) error {

	// Thêm members mới
	for _, memberID := range members {

		// Lưu vào DB
		rosterModel := RosterModel{
			UserID: memberID,
			TeamID: teamID,
			Role:   entity.TeamMember, // Mặc định là MEMBER
		}
		if err := r.db.WithContext(ctx).Create(&rosterModel).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *TeamRepositoryImpl) AddManager(ctx context.Context, teamID uuid.UUID, managerID uuid.UUID) error {
	// Lưu vào DB
	rosterModel := RosterModel{
		UserID: managerID,
		TeamID: teamID,
		Role:   entity.TeamManager, // Mặc định là MANAGER
	}
	if err := r.db.WithContext(ctx).Create(&rosterModel).Error; err != nil {
		return err
	}
	return nil
}

// Remove implements entity.TeamRepository
func (r *TeamRepositoryImpl) RemoveMember(ctx context.Context, teamID uuid.UUID, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("team_id = ? AND user_id = ?", teamID, memberID).Delete(&RosterModel{}).Error
}

// Update implements entity.TeamRepository
func (r *TeamRepositoryImpl) Update(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	model := TeamModelFromDomain(team)
	if err := r.db.WithContext(ctx).Model(&TeamModel{ID: team.ID}).Updates(map[string]interface{}{
		"team_name": model.TeamName,
	}).Error; err != nil {
		return nil, err
	}
	return team, nil
}

// Delete implements entity.TeamRepository
func (r *TeamRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()

	// First delete all associated rosters
	if err := tx.Delete(&RosterModel{}, "team_id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Then delete the team
	if err := tx.Delete(&TeamModel{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
