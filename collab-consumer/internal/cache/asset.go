package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type AssetCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewAssetCache(rdb *redis.Client) *AssetCache {
	return &AssetCache{rdb: rdb, ttl: 24 * time.Hour}
}

func assetACLKey(assetID string) string {
	return fmt.Sprintf("asset:%s:acl", assetID)
}

// Grant or update user access
func (c *AssetCache) SetUserAccess(ctx context.Context, assetID, userID, accessType string) error {
	key := assetACLKey(assetID)
	if err := c.rdb.HSet(ctx, key, userID, accessType).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, c.ttl).Err()
}

// Remove user access
func (c *AssetCache) RemoveUserAccess(ctx context.Context, assetID, userID string) error {
	key := assetACLKey(assetID)
	if err := c.rdb.HDel(ctx, key, userID).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, c.ttl).Err()
}

// Get user access
func (c *AssetCache) GetUserAccess(ctx context.Context, assetID, userID string) (string, error) {
	key := assetACLKey(assetID)
	result, err := c.rdb.HGet(ctx, key, userID).Result()
	if err == nil {
		// Reset TTL on successful access
		c.rdb.Expire(ctx, key, c.ttl)
	}
	return result, err
}

// Get all ACL for an asset
func (c *AssetCache) GetAssetACL(ctx context.Context, assetID string) (map[string]string, error) {
	key := assetACLKey(assetID)
	result, err := c.rdb.HGetAll(ctx, key).Result()
	if err == nil && len(result) > 0 {
		// Reset TTL on successful access
		c.rdb.Expire(ctx, key, c.ttl)
	}
	return result, err
}
