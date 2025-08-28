package cache

import (
	"collab-service/internal/domain/entity"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TeamRepositoryWithCache struct {
	dbRepo entity.TeamRepository
	rdb    *redis.Client
	ttl    time.Duration
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

func NewTeamRepositoryWithCache(dbRepo entity.TeamRepository, rdb *redis.Client, ttl time.Duration) entity.TeamRepository {
	return &TeamRepositoryWithCache{
		dbRepo: dbRepo,
		rdb:    rdb,
		ttl:    ttl,
	}
}

func teamCacheKey(teamID uuid.UUID) string {
	return fmt.Sprintf("team:%s:members", teamID.String())
}

//
// CACHE HELPERS
//

func (r *TeamRepositoryWithCache) cacheTeam(ctx context.Context, team *entity.Team) error {
	memberIDs := make([]interface{}, len(team.Rosters))
	for i, ro := range team.Rosters {
		memberIDs[i] = ro.UserID.String()
	}

	// ghi set members
	if err := r.rdb.SAdd(ctx, teamCacheKey(team.ID), memberIDs...).Err(); err != nil {
		return err
	}
	return nil
}

func (r *TeamRepositoryWithCache) getCachedTeam(ctx context.Context, teamID uuid.UUID) (*entity.Team, error) {
	userIDs, err := r.rdb.SMembers(ctx, teamCacheKey(teamID)).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	rosters := make([]entity.Roster, len(userIDs))
	for i, uid := range userIDs {
		uidParsed, _ := uuid.Parse(uid)
		rosters[i] = entity.Roster{
			UserID: uidParsed,
			TeamID: teamID,
			// không có role
		}
	}

	return &entity.Team{
		ID:      teamID,
		Name:    "", // Name không cache
		Rosters: rosters,
	}, nil
}

//
// IMPLEMENT ENTITY.TEAMREPOSITORY
//

func (r *TeamRepositoryWithCache) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	// thử lấy cache
	if team, err := r.getCachedTeam(ctx, id); err == nil && team != nil {
		log.Println("CACHE HIT")
		return team, nil
	}

	// fallback DB
	team, err := r.dbRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = r.cacheTeam(ctx, team)
	return team, nil
}

func (r *TeamRepositoryWithCache) GetRole(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (entity.TeamAccessRole, error) {
	return r.dbRepo.GetRole(ctx, teamID, userID)
}

func (r *TeamRepositoryWithCache) AddMembers(ctx context.Context, teamID uuid.UUID, members []uuid.UUID) error {
	if err := r.dbRepo.AddMembers(ctx, teamID, members); err != nil {
		return err
	}
	return nil
}

func (r *TeamRepositoryWithCache) AddManager(ctx context.Context, teamID uuid.UUID, managerID uuid.UUID) error {
	if err := r.dbRepo.AddManager(ctx, teamID, managerID); err != nil {
		return err
	}

	return nil
}

func (r *TeamRepositoryWithCache) RemoveMember(ctx context.Context, teamID uuid.UUID, memberID uuid.UUID) error {
	if err := r.dbRepo.RemoveMember(ctx, teamID, memberID); err != nil {
		return err
	}

	return nil
}

// Create implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) Create(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	return r.dbRepo.Create(ctx, team)
}

// Delete implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) Delete(ctx context.Context, id uuid.UUID) error {
	// Invalidate cache before deletion
	_ = r.rdb.Del(ctx, teamCacheKey(id)).Err()
	return r.dbRepo.Delete(ctx, id)
}

// ExistsByID implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.dbRepo.ExistsByID(ctx, id)
}

// GetAllByUserID implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Team, error) {
	return r.dbRepo.GetAllByUserID(ctx, userID)
}

// List implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) List(ctx context.Context) ([]*entity.Team, error) {
	return r.dbRepo.List(ctx)
}

// Update implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) Update(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	// Invalidate cache before update
	return r.dbRepo.Update(ctx, team)
}
