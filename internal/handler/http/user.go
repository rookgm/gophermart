package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/rookgm/gophermart/internal/service"
	"net/http"
	"time"
)

// UserService is interface for interfacing with user-related logic
type UserService interface {
	// Register is registers new user
	Register(ctx context.Context, user *models.User) (*models.User, error)
}

// UserHandler represents HTTP handler for user-related requests
type UserHandler struct {
	userSvc  UserService
	tokenSvc service.TokenService
}

// NewUserHandler creates new UserHandler instance
func NewUserHandler(us UserService) *UserHandler {
	return &UserHandler{
		userSvc: us,
	}
}

// registerRequest is user registration data
type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterUser registers new user
// 200 — пользователь успешно зарегистрирован и аутентифицирован;
// 400 — неверный формат запроса;
// 409 — логин уже занят;
// 500 — внутренняя ошибка сервера.
func (uh *UserHandler) RegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var regReq registerRequest

		if err := json.NewDecoder(r.Body).Decode(&regReq); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		user := models.User{
			Login:    regReq.Login,
			Password: regReq.Password,
		}

		_, err := uh.userSvc.Register(r.Context(), &user)
		if err != nil {
			if errors.Is(err, models.ErrConflictData) {
				http.Error(w, "bad request", http.StatusConflict)
				return
			} else {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		token, err := uh.tokenSvc.CreateToken(&user)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_gophermart",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		w.WriteHeader(http.StatusOK)
	}
}
