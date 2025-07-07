package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rookgm/gophermart/internal/models"
	"net/http"
	"time"
)

// AuthService is interface for interfacing with user authentication
type AuthService interface {
	Login(ctx context.Context, login string, password string) (string, error)
}

// AuthHandler represents HTTP handler for user-related requests
type AuthHandler struct {
	authSvc AuthService
}

// NewAuthHandler creates NewAuthHandler instance
func NewAuthHandler(authSvc AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// loginRequest is user login data
type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginUser perform user logging
// 200 — пользователь успешно аутентифицирован;
// 400 — неверный формат запроса;
// 401 — неверная пара логин/пароль;
// 500 — внутренняя ошибка сервера.
func (ah *AuthHandler) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReq loginRequest

		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		token, err := ah.authSvc.Login(r.Context(), loginReq.Login, loginReq.Password)
		if err != nil {
			if errors.Is(err, models.ErrInvalidCredentials) {
				http.Error(w, "incorrect login or password", http.StatusUnauthorized)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		w.WriteHeader(http.StatusOK)
	}
}
