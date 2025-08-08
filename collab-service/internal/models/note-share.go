package models

type NoteShare struct {
	ID     string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	NoteID string `gorm:"type:uuid;not null"`
	UserID string `gorm:"type:uuid;not null"`

	Access AccessRole
}

func (NoteShare) TableName() string {
	return "note_shares"
}
