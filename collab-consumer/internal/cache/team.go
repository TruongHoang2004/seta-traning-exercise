package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

func (c *Cache) CloseRedis() {
	panic("unimplemented")
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func teamMembersKey(teamID string) string {
	return fmt.Sprintf("team:%s:members", teamID)
}

// Add member to team cache
func (c *Cache) AddTeamMember(ctx context.Context, teamID, userID string) error {
	log.Printf("Adding user %s to team %s cache", userID, teamID)
	return c.rdb.SAdd(ctx, teamMembersKey(teamID), userID).Err()
}

// Remove member from team cache
func (c *Cache) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	return c.rdb.SRem(ctx, teamMembersKey(teamID), userID).Err()
}

// Get all members (cache read)
func (c *Cache) GetTeamMembers(ctx context.Context, teamID string) ([]string, error) {
	return c.rdb.SMembers(ctx, teamMembersKey(teamID)).Result()
}

func folderKey(folderID string) string {
	return fmt.Sprintf("folder:%s", folderID)
}

func noteKey(noteID string) string {
	return fmt.Sprintf("note:%s", noteID)
}

// Set folder metadata
func (c *Cache) SetFolderMetadata(ctx context.Context, folderID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, folderKey(folderID), b, 0).Err()
}

// Get folder metadata
func (c *Cache) GetFolderMetadata(ctx context.Context, folderID string, dest interface{}) error {
	val, err := c.rdb.Get(ctx, folderKey(folderID)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}

// Invalidate folder cache
func (c *Cache) DeleteFolderMetadata(ctx context.Context, folderID string) error {
	return c.rdb.Del(ctx, folderKey(folderID)).Err()
}

// Set note metadata
func (c *Cache) SetNoteMetadata(ctx context.Context, noteID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, noteKey(noteID), b, 0).Err()
}

// Get note metadata
func (c *Cache) GetNoteMetadata(ctx context.Context, noteID string, dest interface{}) error {
	val, err := c.rdb.Get(ctx, noteKey(noteID)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}

// Invalidate note cache
func (c *Cache) DeleteNoteMetadata(ctx context.Context, noteID string) error {
	return c.rdb.Del(ctx, noteKey(noteID)).Err()
}

func assetACLKey(assetID string) string {
	return fmt.Sprintf("asset:%s:acl", assetID)
}

// Grant or update user access
func (c *Cache) SetUserAccess(ctx context.Context, assetID, userID, accessType string) error {
	return c.rdb.HSet(ctx, assetACLKey(assetID), userID, accessType).Err()
}

// Remove user access
func (c *Cache) RemoveUserAccess(ctx context.Context, assetID, userID string) error {
	return c.rdb.HDel(ctx, assetACLKey(assetID), userID).Err()
}

// Get user access
func (c *Cache) GetUserAccess(ctx context.Context, assetID, userID string) (string, error) {
	return c.rdb.HGet(ctx, assetACLKey(assetID), userID).Result()
}

// Get all ACL for an asset
func (c *Cache) GetAssetACL(ctx context.Context, assetID string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, assetACLKey(assetID)).Result()
}
