package models

import "time"

type UserType string

const (
	UserTypeManager UserType = "MANAGER"
	UserTypeMember  UserType = "MEMBER"
)

type User struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username  string    `gorm:"not null"`
	Email     string    `gorm:"email;uniqueIndex"`
	Password  string    `gorm:"not null"`
	Role      UserType  `gorm:"not null;default:MEMBER"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
