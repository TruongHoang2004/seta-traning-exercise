package cache

import (
	"collab-service/internal/domain/entity"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TTL cho metadata
const metadataTTL = 5 * time.Minute

type CachedMember struct {
	UserID uuid.UUID             `json:"user_id"`
	Role   entity.TeamAccessRole `json:"role"`
}

type CacheService struct {
	rdb *redis.Client
}

func NewCacheService(rdb *redis.Client) *CacheService {
	return &CacheService{rdb: rdb}
}

// Team members cache
func (c *CacheService) AddTeamMember(ctx context.Context, teamID uuid.UUID, member CachedMember) error {
	key := "team:" + teamID.String() + ":members"

	data, err := json.Marshal(member)
	if err != nil {
		return err
	}

	if err := c.rdb.HSet(ctx, key, member.UserID.String(), data).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, time.Hour).Err()
}

func (c *CacheService) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	key := "team:" + teamID.String() + ":members"
	if err := c.rdb.HDel(ctx, key, userID.String()).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, time.Hour).Err()
}

func (c *CacheService) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]CachedMember, error) {
	key := "team:" + teamID.String() + ":members"

	values, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	members := make([]CachedMember, 0, len(values))
	for _, v := range values {
		var member CachedMember
		if err := json.Unmarshal([]byte(v), &member); err == nil {
			members = append(members, member)
		}
	}
	if len(members) > 0 {
		_ = c.rdb.Expire(ctx, key, time.Hour).Err()
	}
	return members, nil
}

func (c *CacheService) DeleteTeamMembers(ctx context.Context, teamID uuid.UUID) error {
	key := "team:" + teamID.String() + ":members"
	return c.rdb.Del(ctx, key).Err()
}

// Folder / Note metadata cache
func (c *CacheService) SetFolderMetadata(ctx context.Context, folderID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return c.rdb.Set(ctx, "folder:"+folderID, b, metadataTTL).Err()
}

func (c *CacheService) GetFolderMetadata(ctx context.Context, folderID string, dest interface{}) error {
	val, err := c.rdb.Get(ctx, "folder:"+folderID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *CacheService) SetNoteMetadata(ctx context.Context, noteID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return c.rdb.Set(ctx, "note:"+noteID, b, metadataTTL).Err()
}

func (c *CacheService) GetNoteMetadata(ctx context.Context, noteID string, dest interface{}) error {
	val, err := c.rdb.Get(ctx, "note:"+noteID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// ACL cache
func (c *CacheService) SetAssetACL(ctx context.Context, assetID, userID, accessType string) error {
	return c.rdb.HSet(ctx, "asset:"+assetID+":acl", userID, accessType).Err()
}

func (c *CacheService) RemoveAssetACL(ctx context.Context, assetID, userID string) error {
	return c.rdb.HDel(ctx, "asset:"+assetID+":acl", userID).Err()
}

func (c *CacheService) GetAssetAccess(ctx context.Context, assetID, userID string) (string, error) {
	return c.rdb.HGet(ctx, "asset:"+assetID+":acl", userID).Result()
}
