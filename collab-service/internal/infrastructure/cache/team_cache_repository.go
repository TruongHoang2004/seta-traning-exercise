package cache

import (
	"collab-service/internal/domain/entity"
	"context"
	"encoding/json"
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
	members := make([]CachedMember, len(team.Rosters))
	for i, ro := range team.Rosters {
		members[i] = CachedMember{
			UserID: ro.UserID,
			Role:   ro.Role,
		}
	}
	cached := CachedTeam{
		ID:      team.ID,
		Name:    team.Name,
		Members: members,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, teamCacheKey(team.ID), data, r.ttl).Err()
}

func (r *TeamRepositoryWithCache) getCachedTeam(ctx context.Context, teamID uuid.UUID) (*entity.Team, error) {
	val, err := r.rdb.Get(ctx, teamCacheKey(teamID)).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	var cached CachedTeam
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, err
	}

	// convert về entity.Team
	rosters := make([]entity.Roster, len(cached.Members))
	for i, m := range cached.Members {
		rosters[i] = entity.Roster{
			UserID: m.UserID,
			TeamID: cached.ID,
			Role:   m.Role,
		}
	}

	return &entity.Team{
		ID:      cached.ID,
		Name:    cached.Name,
		Rosters: rosters,
	}, nil
}

func (r *TeamRepositoryWithCache) addUserToCachedTeam(ctx context.Context, teamID uuid.UUID, userIDs []uuid.UUID, role entity.TeamAccessRole) error {
	team, err := r.getCachedTeam(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return nil // cache miss
	}

	// Check if users already in team and add only new ones
	for _, userID := range userIDs {
		userExists := false
		for _, m := range team.Rosters {
			if m.UserID == userID {
				userExists = true
				break
			}
		}

		if !userExists {
			// Add new user to cached team
			team.Rosters = append(team.Rosters, entity.Roster{
				UserID: userID,
				TeamID: teamID,
				Role:   role,
			})
		}
	}

	return r.cacheTeam(ctx, team)
}

func (r *TeamRepositoryWithCache) removeUserToCachedTeam(ctx context.Context, teamID, userID uuid.UUID) error {
	team, err := r.getCachedTeam(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return nil // cache miss
	}

	// remove from cached team
	newRosters := make([]entity.Roster, 0, len(team.Rosters))
	for _, m := range team.Rosters {
		if m.UserID != userID {
			newRosters = append(newRosters, m)
		}
	}
	team.Rosters = newRosters

	return r.cacheTeam(ctx, team)
}

func (r *TeamRepositoryWithCache) invalidateCache(ctx context.Context, teamID uuid.UUID) {
	_ = r.rdb.Del(ctx, teamCacheKey(teamID)).Err()
}

//
// IMPLEMENT ENTITY.TEAMREPOSITORY
//

func (r *TeamRepositoryWithCache) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	// thử lấy cache
	if team, err := r.getCachedTeam(ctx, id); err == nil && team != nil {
		log.Println("Cache hit")
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

	r.addUserToCachedTeam(ctx, teamID, members, entity.TeamMember)

	// Fallback to invalidation if we couldn't update the cache
	r.invalidateCache(ctx, teamID)
	return nil
}

func (r *TeamRepositoryWithCache) AddManager(ctx context.Context, teamID uuid.UUID, managerID uuid.UUID) error {
	if err := r.dbRepo.AddManager(ctx, teamID, managerID); err != nil {
		return err
	}

	r.addUserToCachedTeam(ctx, teamID, []uuid.UUID{managerID}, entity.TeamManager)

	// Fallback to invalidation if we couldn't update the cache
	r.invalidateCache(ctx, teamID)
	return nil
}

func (r *TeamRepositoryWithCache) RemoveMember(ctx context.Context, teamID uuid.UUID, memberID uuid.UUID) error {
	if err := r.dbRepo.RemoveMember(ctx, teamID, memberID); err != nil {
		return err
	}

	r.removeUserToCachedTeam(ctx, teamID, memberID)

	r.invalidateCache(ctx, teamID)
	return nil
}

// Create implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) Create(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	return r.dbRepo.Create(ctx, team)
}

// Delete implements entity.TeamRepository.
func (r *TeamRepositoryWithCache) Delete(ctx context.Context, id uuid.UUID) error {
	// Invalidate cache before deletion
	r.invalidateCache(ctx, id)
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
	r.invalidateCache(ctx, team.ID)
	return r.dbRepo.Update(ctx, team)
}
