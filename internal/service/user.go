package service

import (
	"context"
	"github.com/rookgm/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository is interface for interacting with user-related data
type UserRepository interface {
	// CreateUser insert new user into database
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	// GetUserByLogin retrieves user info by login
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

// UserService implements UserService interface
type UserService struct {
	repo UserRepository
	//tokenSvc TokenService
}

// NewUserService creates new UserService instance
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register is registers new user
func (us *UserService) Register(ctx context.Context, user *models.User) error {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	user, err = us.repo.CreateUser(ctx, user)

	return err
}

// HashPassword returns bcrypt hash of password
func HashPassword(password string) (string, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashPwd), nil
}
