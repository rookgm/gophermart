package service

import "github.com/rookgm/gophermart/internal/models"

type TokenService interface {
	CreateToken(user *models.User) (string, error)
	VerifyToken(tokenString string) (*models.TokenPayload, error)
}
