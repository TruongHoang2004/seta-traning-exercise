package models

type Team struct {
	ID       string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TeamName string
	Rosters  []Roster `gorm:"foreignKey:TeamID"`
}
