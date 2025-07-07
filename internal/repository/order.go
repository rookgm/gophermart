package repository

import (
	"context"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/rookgm/gophermart/internal/repository/postgres"
)

// OrderRepository implements OrderRepository interface
type OrderRepository struct {
	db *postgres.DB
}

// NewOrderRepository creates new OrderRepository instance
func NewOrderRepository(db *postgres.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder inserts new order to database
func (or *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	return order, nil
}
