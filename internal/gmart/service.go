package gmart

import "github.com/rookgm/gophermart/config"

type GMartRepository interface {
}

type Service struct {
	repo GMartRepository
	cfg  *config.Config
}

func NewService(repo GMartRepository, cfg *config.Config) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}
