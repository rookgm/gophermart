package handler

import (
	"go.uber.org/zap"
	"net/http"
)

type GMartService interface {
}

type Handler struct {
	service GMartService
	logger  *zap.Logger
}

func New(service GMartService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

//POST /api/user/register HTTP/1.1
//Content-Type: application/json
//...
//
//{
//"login": "<login>",
//"password": "<password>"
//}

func (h *Handler) RegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}
}

func (h *Handler) AuthenticationUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
