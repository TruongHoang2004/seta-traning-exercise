package repository

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/persistence/model"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.TeamModel{}, &model.RosterModel{})
	require.NoError(t, err)

	return db
}

func createTestTeam(name string) *entity.Team {
	return &entity.Team{
		Name:      name,
		Rosters:   []entity.Roster{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestTeamRepository_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTeamRepository(db)

	ctx := context.Background()
	team := createTestTeam("Team A")

	saved, err := repo.Create(ctx, team)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, saved.ID)

	got, err := repo.GetByID(ctx, saved.ID)
	require.NoError(t, err)
	require.Equal(t, saved.ID, got.ID)
	require.Equal(t, saved.Name, got.Name)
}

func TestTeamRepository_AddMembersAndGetRole(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTeamRepository(db)

	ctx := context.Background()
	team := createTestTeam("Team B")
	saved, _ := repo.Create(ctx, team)

	memberID := uuid.New()
	err := repo.AddMembers(ctx, saved.ID, []uuid.UUID{memberID})
	require.NoError(t, err)

	role, err := repo.GetRole(ctx, saved.ID, memberID)
	require.NoError(t, err)
	require.Equal(t, entity.TeamMember, role)
}
func TestTeamRepository_AddManager(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTeamRepository(db)

	ctx := context.Background()
	team := createTestTeam("Team C")
	saved, _ := repo.Create(ctx, team)

	managerID := uuid.New()
	err := repo.AddManager(ctx, saved.ID, managerID)
	require.NoError(t, err)

	role, err := repo.GetRole(ctx, saved.ID, managerID)
	require.NoError(t, err)
	require.Equal(t, entity.TeamManager, role)
}
func TestTeamRepository_RemoveMember(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTeamRepository(db)

	ctx := context.Background()
	team := createTestTeam("Team D")
	saved, _ := repo.Create(ctx, team)

	memberID := uuid.New()
	repo.AddMembers(ctx, saved.ID, []uuid.UUID{memberID})

	err := repo.RemoveMember(ctx, saved.ID, memberID)
	require.NoError(t, err)

	role, err := repo.GetRole(ctx, saved.ID, memberID)
	require.NoError(t, err)
	require.Equal(t, entity.TeamNone, role)
}
func TestTeamRepository_GetAllByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTeamRepository(db)
	ctx := context.Background()

	team1 := createTestTeam("Team1")
	team2 := createTestTeam("Team2")
	saved1, _ := repo.Create(ctx, team1)
	saved2, _ := repo.Create(ctx, team2)

	userID := uuid.New()
	repo.AddMembers(ctx, saved1.ID, []uuid.UUID{userID})
	repo.AddMembers(ctx, saved2.ID, []uuid.UUID{userID})

	teams, err := repo.GetAllByUserID(ctx, userID)
	require.NoError(t, err)
	require.Len(t, teams, 2)
}
