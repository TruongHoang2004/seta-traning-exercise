package models

type Roster struct {
	ID     string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID string `gorm:"type:uuid"`
	User   *User  `gorm:"foreignKey:UserID;references:ID"`

	TeamID string `gorm:"type:uuid"`
	Team   *Team  `gorm:"foreignKey:TeamID;references:ID"`

	Role     UserRole `gorm:"type:varchar(20)"`
	IsLeader bool     `gorm:"default:false"`
}
