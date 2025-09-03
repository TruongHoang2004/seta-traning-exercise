package user_service

import (
	"collab-service/internal/domain/entity"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	UserID    uuid.UUID       `json:"userId"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	Role      entity.UserType `json:"role"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (m *UserModel) ToDomain() *entity.User {
	return &entity.User{
		ID:        m.UserID,
		Username:  m.Username,
		Email:     m.Email,
		Role:      m.Role,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserModelFromDomain(domainUser *entity.User) *UserModel {
	return &UserModel{
		UserID:    domainUser.ID,
		Username:  domainUser.Username,
		Email:     domainUser.Email,
		Password:  domainUser.Password,
		Role:      domainUser.Role,
		CreatedAt: domainUser.CreatedAt,
		UpdatedAt: domainUser.UpdatedAt,
	}
}

func NewUserRepository(client *GraphQLClient) entity.UserRepository {
	return &UserRepositoryImpl{
		client: client,
	}
}

type UserRepositoryImpl struct {
	client *GraphQLClient
}

// Create implements entity.UserRepository.
func (u *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	input := CreateUserInput{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}

	createdUser, err := u.client.CreateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	return createdUser.User.ToDomain(), nil
}

func (u *UserRepositoryImpl) CreateMany(ctx context.Context, users []*entity.User) ([]*entity.User, []error) {
	var wg sync.WaitGroup
	usersChan := make(chan *entity.User, len(users))
	errorsChan := make(chan error, len(users))

	for _, user := range users {
		wg.Add(1)
		go func(user *entity.User) {
			defer wg.Done()
			createdUser, err := u.Create(ctx, user)
			if err != nil {
				errorsChan <- err
				return
			}
			usersChan <- createdUser
		}(user)
	}

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(usersChan)
		close(errorsChan)
	}()

	// Collect results and errors
	var createdUsers []*entity.User
	var errors []error

	// Collect all users first
	for user := range usersChan {
		createdUsers = append(createdUsers, user)
	}

	// Then collect all errors
	for err := range errorsChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	return createdUsers, errors
}

// Delete implements entity.UserRepository.
func (u *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

func (u *UserRepositoryImpl) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	// ExistsByID checks if a user with the given ID exists
	user, err := u.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// GetByID implements entity.UserRepository.
func (u *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := u.client.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// List implements entity.UserRepository.
func (u *UserRepositoryImpl) List(ctx context.Context, userType *entity.UserType, userIDs []uuid.UUID) ([]*entity.User, error) {
	// Convert UUIDs to strings
	userIDStrings := make([]string, len(userIDs))
	for i, id := range userIDs {
		userIDStrings[i] = id.String()
	}

	// Pass nil for UserType if you want to get users of any type
	users, err := u.client.GetUsers(ctx, userType, userIDStrings)
	if err != nil {
		return nil, err
	}

	// Convert []UserModel to []*entity.User
	domainUsers := make([]*entity.User, len(users))
	for i, user := range users {
		domainUsers[i] = user.ToDomain()
	}

	return domainUsers, nil
}

// Update implements entity.UserRepository.
func (u *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	panic("unimplemented")
}
