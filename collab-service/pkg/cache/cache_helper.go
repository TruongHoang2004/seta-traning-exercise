package cache

import (
	"context"
	"encoding/json"
	"time"
)

// TTL cho metadata (tùy chọn)
const metadataTTL = 5 * time.Minute

// Team members cache
func GetTeamMembers(ctx context.Context, teamID string) ([]string, error) {
	return Rdb.SMembers(ctx, "team:"+teamID+":members").Result()
}
func AddTeamMember(ctx context.Context, teamID, userID string) error {
	return Rdb.SAdd(ctx, "team:"+teamID+":members", userID).Err()
}
func RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	return Rdb.SRem(ctx, "team:"+teamID+":members", userID).Err()
}

// Folder / Note metadata cache
func SetFolderMetadata(ctx context.Context, folderID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return Rdb.Set(ctx, "folder:"+folderID, b, metadataTTL).Err()
}
func GetFolderMetadata(ctx context.Context, folderID string, dest interface{}) error {
	val, err := Rdb.Get(ctx, "folder:"+folderID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}
func SetNoteMetadata(ctx context.Context, noteID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return Rdb.Set(ctx, "note:"+noteID, b, metadataTTL).Err()
}
func GetNoteMetadata(ctx context.Context, noteID string, dest interface{}) error {
	val, err := Rdb.Get(ctx, "note:"+noteID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// ACL cache
func SetAssetACL(ctx context.Context, assetID, userID, accessType string) error {
	return Rdb.HSet(ctx, "asset:"+assetID+":acl", userID, accessType).Err()
}
func RemoveAssetACL(ctx context.Context, assetID, userID string) error {
	return Rdb.HDel(ctx, "asset:"+assetID+":acl", userID).Err()
}
func GetAssetAccess(ctx context.Context, assetID, userID string) (string, error) {
	return Rdb.HGet(ctx, "asset:"+assetID+":acl", userID).Result()
}
