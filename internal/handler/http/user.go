package http

import "github.com/rookgm/gophermart/internal/service"

type UserHandler struct {
	us service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{
		us: us,
	}
}
