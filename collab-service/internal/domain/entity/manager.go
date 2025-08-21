package entity

import (
	"context"

	"github.com/google/uuid"
)

type ManagerRepository interface {
	GetAssetsByTeamID(ctx context.Context, teamID uuid.UUID) ([]*Note, error)
	GetAssetsByUserID(ctx context.Context, userID uuid.UUID) ([]*Note, error)
}
