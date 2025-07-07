package service

import (
	"context"
	"errors"
	"github.com/rookgm/gophermart/internal/auth"
	"github.com/rookgm/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthService implements AuthService interface
type AuthService struct {
	repo     UserRepository
	tokenSvc TokenService
}

// NewAuthService creates AuthService instance
func NewAuthService(repo UserRepository, ts TokenService) *AuthService {
	return &AuthService{repo: repo, tokenSvc: ts}
}

// Login authenticates registered user
func (as *AuthService) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := as.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, models.ErrDataNotFound) {
			return "", models.ErrInvalidCredentials
		}
		return "", err
	}

	if err := ComparePassword(password, user.Password); err != nil {
		return "", models.ErrInvalidCredentials
	}

	token, err := as.tokenSvc.CreateToken(user)
	if err != nil {
		return "", auth.ErrTokenCreate
	}

	return token, nil
}

// ComparePassword compares a bcrypt hashed password with its possible plaintext equivalent.
// Returns nil on success, or an error on failure.
func ComparePassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
