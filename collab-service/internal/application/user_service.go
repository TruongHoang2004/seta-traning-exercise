package application

import (
	"collab-service/internal/domain/entity"
	"context"
)

type UserService struct {
	repo entity.UserRepository
}

func NewUserService(repo entity.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateManyUsers(ctx context.Context, users []*entity.User, password string) ([]*entity.User, []error) {
	return s.repo.CreateMany(ctx, users, password)
}
