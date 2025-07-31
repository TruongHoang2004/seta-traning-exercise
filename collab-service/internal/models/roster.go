package models

type Roster struct {
	ID     string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID string `gorm:"type:uuid"`

	TeamID string `gorm:"type:uuid"`
	Team   *Team  `gorm:"foreignKey:TeamID;references:ID"`

	IsLeader bool `gorm:"default:false"`
}
