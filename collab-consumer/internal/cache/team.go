package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type TeamCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func (c *TeamCache) CloseRedis() {
	if c.rdb != nil {
		c.rdb.Close()
	}
}

func NewCache(rdb *redis.Client) *TeamCache {
	return &TeamCache{rdb: rdb, ttl: time.Hour}
}

func teamMembersKey(teamID string) string {
	return fmt.Sprintf("team:%s:members", teamID)
}

// Add member to team cache
func (c *TeamCache) AddTeamMember(ctx context.Context, teamID, userID string) error {
	log.Printf("Adding user %s to team %s cache", userID, teamID)
	key := teamMembersKey(teamID)
	if err := c.rdb.SAdd(ctx, key, userID).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, c.ttl).Err()
}

// Remove member from team cache
func (c *TeamCache) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	return c.rdb.SRem(ctx, teamMembersKey(teamID), userID).Err()
}

// Get all members (cache read)
func (c *TeamCache) GetTeamMembers(ctx context.Context, teamID string) ([]string, error) {
	key := teamMembersKey(teamID)
	members, err := c.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	// Refresh TTL on read
	c.rdb.Expire(ctx, key, c.ttl)
	return members, nil
}

func folderKey(folderID string) string {
	return fmt.Sprintf("folder:%s", folderID)
}

func noteKey(noteID string) string {
	return fmt.Sprintf("note:%s", noteID)
}

// Set folder metadata
func (c *TeamCache) SetFolderMetadata(ctx context.Context, folderID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, folderKey(folderID), b, c.ttl).Err()
}

// Get folder metadata
func (c *TeamCache) GetFolderMetadata(ctx context.Context, folderID string, dest interface{}) error {
	key := folderKey(folderID)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	// Refresh TTL on read
	c.rdb.Expire(ctx, key, c.ttl)
	return json.Unmarshal(val, dest)
}

// Invalidate folder cache
func (c *TeamCache) DeleteFolderMetadata(ctx context.Context, folderID string) error {
	return c.rdb.Del(ctx, folderKey(folderID)).Err()
}

// Set note metadata
func (c *TeamCache) SetNoteMetadata(ctx context.Context, noteID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, noteKey(noteID), b, c.ttl).Err()
}

// Get note metadata
func (c *TeamCache) GetNoteMetadata(ctx context.Context, noteID string, dest interface{}) error {
	key := noteKey(noteID)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	// Refresh TTL on read
	c.rdb.Expire(ctx, key, c.ttl)
	return json.Unmarshal(val, dest)
}

// Invalidate note cache
func (c *TeamCache) DeleteNoteMetadata(ctx context.Context, noteID string) error {
	return c.rdb.Del(ctx, noteKey(noteID)).Err()
}
