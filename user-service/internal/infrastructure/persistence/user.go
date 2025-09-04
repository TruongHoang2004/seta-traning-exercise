package persistence

import (
	"time"
	"user-service/internal/graphql/model"
)

type User struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username  string         `gorm:"not null"`
	Email     string         `gorm:"email;uniqueIndex"`
	Password  string         `gorm:"not null"`
	Role      model.UserType `gorm:"not null;default:MEMBER"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

// ToModel converts a persistence User to a domain model User
func (u *User) ToModel() *model.User {
	createdAtStr := u.CreatedAt.Format(time.RFC3339)
	return &model.User{
		UserID:    u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: &createdAtStr,
	}
}

// FromModel converts a domain model User to a persistence User
func FromModel(m *model.User) *User {
	var createdAt time.Time
	if m.CreatedAt != nil {
		createdAt, _ = time.Parse(time.RFC3339, *m.CreatedAt)
	}
	return &User{
		ID:        m.UserID,
		Username:  m.Username,
		Email:     m.Email,
		Role:      m.Role,
		CreatedAt: createdAt,
		// UpdatedAt will be set automatically by GORM
	}
}
