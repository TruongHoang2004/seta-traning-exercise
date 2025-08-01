package models

type Note struct {
	ID string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

	Title string `gorm:"type:varchar(255);not null"`
	Body  string `gorm:"type:text;not null"`

	FolderID string  `gorm:"type:uuid;not null"`
	Folder   *Folder `gorm:"foreignKey:FolderID;references:ID"`

	OwnerID string `gorm:"type:uuid;not null"`
}

func (Note) TableName() string {
	return "notes"
}
