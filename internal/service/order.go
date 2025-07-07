package service

import (
	"context"
	"errors"
	"github.com/phedde/luhn-algorithm"
	"github.com/rookgm/gophermart/internal/models"
	"strconv"
)

// OrderRepository is interface for interacting with order-related data
type OrderRepository interface {
	// CreateOrder inserts new order to database
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
}

// OrderService implements OrderService interface
type OrderService struct {
	repo OrderRepository
}

// NewOrderService creates new NewOrderService instance
func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// Upload uploads user order
func (os *OrderService) Upload(ctx context.Context, order *models.Order) (*models.Order, error) {
	num, err := strconv.ParseInt(order.OrderID, 10, 64)
	if err != nil {
		return nil, err
	}
	// check order id using Luhn algorithm
	if ok := luhn.IsValid(num); !ok {
		return nil, models.ErrInvalidOrderID
	}

	order, err = os.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, models.ErrConflictData) {
			return nil, err
		}
	}

	return order, nil
}
