package repository

import (
	"collab-service/internal/domain/entity"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ManagerRepositoryImpl struct {
	db *gorm.DB
}

func NewManagerRepository(db *gorm.DB) *ManagerRepositoryImpl {
	return &ManagerRepositoryImpl{
		db: db,
	}
}

// GetAssetsByTeamID retrieves all notes that are accessible by members of a team
func (r *ManagerRepositoryImpl) GetAssetsByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Note, error) {
	var notes []*entity.Note

	// Find notes owned by team members
	ownedNotes := r.db.WithContext(ctx).
		Table("notes n").
		Select("n.*").
		Joins("JOIN team_members tm ON n.owner_id = tm.user_id").
		Where("tm.team_id = ?", teamID)

	// Find notes shared with the team
	sharedNotes := r.db.WithContext(ctx).
		Table("notes n").
		Select("n.*").
		Joins("JOIN note_shares ns ON n.id = ns.note_id").
		Joins("JOIN team_members tm ON ns.team_id = tm.team_id").
		Where("tm.team_id = ?", teamID)

	// Combine the queries
	err := r.db.WithContext(ctx).
		Table("(?) UNION (?)", ownedNotes, sharedNotes).
		Scan(&notes).Error

	if err != nil {
		return nil, err
	}

	return notes, nil
}

// GetAssetsByUserID retrieves all notes owned by or shared with the user
func (r *ManagerRepositoryImpl) GetAssetsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Note, error) {
	var notes []*entity.Note

	// Find notes owned by the user
	ownedNotes := r.db.WithContext(ctx).
		Table("notes n").
		Select("DISTINCT n.*").
		Joins("LEFT JOIN note_shares ns ON n.id = ns.note_id").
		Where("n.owner_id = ?", userID)

	// Find notes directly shared with the user
	directlyShared := r.db.WithContext(ctx).
		Table("notes n").
		Select("n.*").
		Joins("JOIN note_shares ns ON n.id = ns.note_id").
		Where("ns.user_id = ?", userID)

	// Find notes shared with teams the user belongs to
	teamShared := r.db.WithContext(ctx).
		Table("notes n").
		Select("n.*").
		Joins("JOIN note_shares ns ON n.id = ns.note_id").
		Joins("JOIN team_members tm ON ns.team_id = tm.team_id").
		Where("tm.user_id = ?", userID)

	// Combine the queries
	err := r.db.WithContext(ctx).
		Table("(?) UNION (?) UNION (?)", ownedNotes, directlyShared, teamShared).
		Scan(&notes).Error

	if err != nil {
		return nil, err
	}

	return notes, nil
}
