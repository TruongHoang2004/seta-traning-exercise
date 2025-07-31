package models

type Folder struct {
	ID string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

	Name    string `gorm:"type:varchar(255);not null"`
	OwnerID string `gorm:"type:uuid;not null"`

	Notes []Note `gorm:"foreignKey:FolderID;references:ID"`
}
