package persistence

import (
	"user-service/internal/graphql/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *model.CreateUserInput) (model.User, error) {
	var userModel = &User{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Password: user.Password,
	}
	if err := r.db.Create(userModel).Error; err != nil {
		return model.User{}, err
	}
	return *userModel.ToModel(), nil
}

func (r *UserRepository) IsExistingEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) GetUserByID(id string) (*model.User, error) {
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return user.ToModel(), nil
}

func (r *UserRepository) GetUserPasswordByEmail(email string) (*model.User, string, error) {
	var user User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, "", err
	}
	return user.ToModel(), user.Password, nil
}

func (r *UserRepository) GetUserByRole(role model.UserType) ([]*model.User, error) {
	var users []*User
	if err := r.db.Where("role = ?", role).Find(&users).Error; err != nil {
		return nil, err
	}
	var result []*model.User
	for _, u := range users {
		result = append(result, u.ToModel())
	}
	return result, nil
}

func (r *UserRepository) QueryUsers(role *model.UserType, userIDs []string) ([]*model.User, error) {
	var users []*User
	query := r.db.Model(&User{})
	if role != nil {
		query = query.Where("role = ?", *role)
	}
	if len(userIDs) > 0 {
		query = query.Where("id IN ?", userIDs)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	var result []*model.User
	for _, u := range users {
		result = append(result, u.ToModel())
	}
	return result, nil
}

func (r *UserRepository) GetAllUsers() ([]*model.User, error) {
	var users []*User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	var result []*model.User
	for _, u := range users {
		result = append(result, u.ToModel())
	}
	return result, nil
}

func (r *UserRepository) UpdateUser(user *model.User) (model.User, error) {
	var userModel = &User{
		ID:       user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}
	if err := r.db.Model(&User{}).Where("id = ?", user.UserID).Updates(userModel).Error; err != nil {
		return model.User{}, err
	}
	return *userModel.ToModel(), nil
}

func (r *UserRepository) DeleteUser(id string) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}
