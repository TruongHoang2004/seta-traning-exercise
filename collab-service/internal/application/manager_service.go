package application

import (
	"collab-service/internal/domain/entity"
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrTeamNotFound = errors.New("team not found")
	ErrUserNotFound = errors.New("user not found")
)

type ManagerService struct {
	assetRepo entity.ManagerRepository
	teamRepo  entity.TeamRepository
	userRepo  entity.UserRepository
}

func NewManagerService(assetRepo entity.ManagerRepository, teamRepo entity.TeamRepository, userRepo entity.UserRepository) *ManagerService {
	return &ManagerService{
		assetRepo: assetRepo,
		teamRepo:  teamRepo,
		userRepo:  userRepo,
	}
}

// GetTeamAssets retrieves all assets that team members own or can access
func (s *ManagerService) GetTeamAssets(ctx context.Context, teamID string) ([]*entity.Note, error) {
	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		return nil, err
	}

	// Verify team exists
	exists, err := s.teamRepo.ExistsByID(ctx, teamUUID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTeamNotFound
	}

	// Get assets for the team
	assets, err := s.assetRepo.GetAssetsByTeamID(ctx, teamUUID)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

// GetUserAssets retrieves all assets owned by or shared with the user
func (s *ManagerService) GetUserAssets(ctx context.Context, userID string) ([]*entity.Note, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Verify user exists
	exists, err := s.userRepo.ExistsByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// Get assets for the user
	assets, err := s.assetRepo.GetAssetsByUserID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return assets, nil
}
