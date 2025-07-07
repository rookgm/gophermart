package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rookgm/gophermart/internal/models"
	"time"
)

var (
	ErrTokenSigningMethod = errors.New("token signing method")
	ErrInvalidToken       = errors.New("token is not valid")
	ErrTokenPayload       = errors.New("token payload is not valid")
	ErrTokenCreate        = errors.New("can not create auth token")
)

type AuthToken struct {
	key []byte
}

func NewAuthToken(key []byte) *AuthToken {
	return &AuthToken{key: key}
}

// CreateToken creates new user token
func (at *AuthToken) CreateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uuid":   uuid.New().String(),
			"userid": user.ID,
			"exp":    time.Now().Add(24 * time.Hour).Unix(),
		})

	return token.SignedString(at.key)
}

// VerifyToken verifies token and return payload
func (at *AuthToken) VerifyToken(tokenString string) (*models.TokenPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenSigningMethod
		}
		return at.key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenPayload
	}

	// get uuid string
	s, ok := claims["uuid"].(string)
	if !ok {
		return nil, ErrTokenPayload
	}

	// parse uuid string
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, ErrTokenPayload
	}

	// get userid in float value
	userIDFloat, ok := claims["userid"].(float64)
	if !ok {
		return nil, ErrTokenPayload
	}

	// convert userid to int
	userID := uint64(userIDFloat)

	return &models.TokenPayload{
		ID:     id,
		UserID: userID,
	}, nil
}
