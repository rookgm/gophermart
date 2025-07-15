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
	// GetOrdersByUserID gets user orders
	GetOrdersByUserID(ctx context.Context, userID uint64) ([]models.Order, error)
	// GetOrderByNumber returns order by number
	GetOrderByNumber(ctx context.Context, num string) (*models.Order, error)
}

// OrderService implements OrderService interface
type OrderService struct {
	repo OrderRepository
}

// NewOrderService creates new OrderService instance
func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// Upload uploads user order
func (os *OrderService) Upload(ctx context.Context, order *models.Order) (*models.Order, error) {
	num, err := strconv.ParseInt(order.Number, 10, 64)
	if err != nil {
		return nil, err
	}
	// check order id using Luhn algorithm
	if ok := luhn.IsValid(num); !ok {
		return nil, models.ErrInvalidOrderNumber
	}

	// check existing order
	curOrder, err := os.repo.GetOrderByNumber(ctx, order.Number)
	if err == nil {
		if curOrder.UserID == order.UserID {
			// order has been loaded by user
			return nil, models.ErrOrderLoadedUser
		}
		// order has been loaded by another user
		return nil, models.ErrOrderLoadedAnotherUser
	}

	// set order status
	order.Status = models.OrderStatusNew

	order, err = os.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, models.ErrConflictData) {
			return nil, err
		}
		return nil, err
	}

	return order, nil
}

// ListUserOrders returns list of user orders
func (os *OrderService) ListUserOrders(ctx context.Context, userID uint64) ([]models.Order, error) {
	return os.repo.GetOrdersByUserID(ctx, userID)
}
