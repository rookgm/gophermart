package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/rookgm/gophermart/internal/repository/postgres"
)

const (
	insertUserQuery = `
					INSERT INTO users (login, password) 
					values ($1, $2)
					RETURNING id, login, password, created_at;
`

	selectUserByLoginQuery = `
					SELECT id, login, password, created_at FROM users
					WHERE login = $1
`
)

// UserRepository implements user repository interface
type UserRepository struct {
	db *postgres.DB
}

// NewUserRepository creates new user repository instance
func NewUserRepository(db *postgres.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser insert new user into database
func (ur *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := ur.db.QueryRow(ctx, insertUserQuery, user.Login, user.Password).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		if errCode := ur.db.ErrorCode(err); errCode == "23505" {
			return nil, models.ErrConflictData
		}
		return nil, err
	}

	return user, nil
}

// GetUserByLogin returns user by login
func (ur *UserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user := models.User{}
	err := ur.db.QueryRow(ctx, selectUserByLoginQuery, login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrDataNotFound
		}
		return nil, err
	}

	return &user, nil
}
