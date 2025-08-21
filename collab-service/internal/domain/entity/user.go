package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserType string

const (
	UserTypeManager UserType = "MANAGER"
	UserTypeMember  UserType = "MEMBER"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Password  string
	Role      UserType
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User, password string) (*User, error)
	CreateMany(ctx context.Context, users []*User, password string) ([]*User, []error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, userType *UserType, userIDs []uuid.UUID) ([]*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
}
