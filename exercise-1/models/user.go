package models

type UserRole string

const (
	RoleManager UserRole = "MANAGER"
	RoleMember  UserRole = "MEMBER"
)

type User struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string
	Email        string `gorm:"unique"`
	PasswordHash string
	Role         UserRole

	Rosters []Roster `gorm:"foreignKey:UserID"`
	Folders []Folder `gorm:"foreignKey:OwnerID"`
}
