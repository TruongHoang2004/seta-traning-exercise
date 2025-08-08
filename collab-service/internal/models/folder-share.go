package models

type AccessRole string

const (
	AccessLevelRead  AccessRole = "READ"
	AccessLevelWrite AccessRole = "WRITE"
)

type FolderShare struct {
	ID       string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FolderID string `gorm:"type:uuid;not null"`
	UserID   string `gorm:"type:uuid;not null"`

	Access AccessRole
}

func (FolderShare) TableName() string {
	return "folder_shares"
}
