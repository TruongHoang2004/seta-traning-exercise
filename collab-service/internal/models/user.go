package models

type UserRole string

const (
	RoleManager UserRole = "MANAGER"
	RoleMember  UserRole = "MEMBER"
)

type User struct {
	ID       string   `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username string   `gorm:"unique;not null"`
	Email    string   `gorm:"unique;not null"`
	Password string   `gorm:"not null"`
	Role     UserRole `gorm:"not null"`
}
