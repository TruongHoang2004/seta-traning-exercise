package entity

import (
	"time"

	"github.com/google/uuid"
)

type AccessLevel string

const (
	AccessLevelOwner AccessLevel = "OWNER"
	AccessLevelRead  AccessLevel = "READ"
	AccessLevelWrite AccessLevel = "WRITE"
	AccessLevelNone  AccessLevel = "NONE"
)

type FolderShare struct {
	ID       uuid.UUID
	FolderID uuid.UUID
	UserID   uuid.UUID

	AccessLevel AccessLevel

	CreatedAt time.Time
}

// Priority trả ra mức độ ưu tiên số của quyền
func (a AccessLevel) Priority() int {
	switch a {
	case AccessLevelOwner:
		return 3
	case AccessLevelWrite:
		return 2
	case AccessLevelRead:
		return 1
	default:
		return 0 // NONE
	}
}

// GreaterThan trả true nếu a có quyền cao hơn b
func (a AccessLevel) GreaterThan(b AccessLevel) bool {
	return a.Priority() >= b.Priority()
}

// MaxAccessLevel trả về quyền cao nhất giữa hai mức
func MaxAccessLevel(a, b AccessLevel) AccessLevel {
	if a.Priority() >= b.Priority() {
		return a
	}
	return b
}
