package cache

import (
	"context"
	"encoding/json"
	"time"
)

// TTL cho metadata (tùy chọn)
const metadataTTL = 5 * time.Minute

// Team members cache
func AddTeamMember(ctx context.Context, teamID, userID string) error {
	key := "team:" + teamID + ":members"
	err := rdb.SAdd(ctx, key, userID).Err()
	if err != nil {
		return err
	}
	// Đặt TTL
	return rdb.Expire(ctx, key, time.Hour).Err()
}

func RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	key := "team:" + teamID + ":members"
	err := rdb.SRem(ctx, key, userID).Err()
	if err != nil {
		return err
	}
	// Reset lại TTL (tùy chiến lược)
	return rdb.Expire(ctx, key, time.Hour).Err()
}

func GetTeamMembers(ctx context.Context, teamID string) ([]string, error) {
	key := "team:" + teamID + ":members"
	return rdb.SMembers(ctx, key).Result()
}

// Folder / Note metadata cache
func SetFolderMetadata(ctx context.Context, folderID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return rdb.Set(ctx, "folder:"+folderID, b, metadataTTL).Err()
}
func GetFolderMetadata(ctx context.Context, folderID string, dest interface{}) error {
	val, err := rdb.Get(ctx, "folder:"+folderID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}
func SetNoteMetadata(ctx context.Context, noteID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return rdb.Set(ctx, "note:"+noteID, b, metadataTTL).Err()
}
func GetNoteMetadata(ctx context.Context, noteID string, dest interface{}) error {
	val, err := rdb.Get(ctx, "note:"+noteID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// ACL cache
func SetAssetACL(ctx context.Context, assetID, userID, accessType string) error {
	return rdb.HSet(ctx, "asset:"+assetID+":acl", userID, accessType).Err()
}
func RemoveAssetACL(ctx context.Context, assetID, userID string) error {
	return rdb.HDel(ctx, "asset:"+assetID+":acl", userID).Err()
}
func GetAssetAccess(ctx context.Context, assetID, userID string) (string, error) {
	return rdb.HGet(ctx, "asset:"+assetID+":acl", userID).Result()
}
