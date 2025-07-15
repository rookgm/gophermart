package service

import (
	"context"
	"errors"
	"github.com/phedde/luhn-algorithm"
	"github.com/rookgm/gophermart/internal/accrual"
	"github.com/rookgm/gophermart/internal/models"
	"strconv"
	"strings"
	"time"
)

// OrderRepository is interface for interacting with order-related data
type OrderRepository interface {
	// CreateOrder inserts new order to database
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	// GetOrdersByUserID gets user orders
	GetOrdersByUserID(ctx context.Context, userID uint64) ([]models.Order, error)
	// GetOrderByNumber returns order by number
	GetOrderByNumber(ctx context.Context, num string) (*models.Order, error)
	// UpdateOrderStatus update order status and accrual
	UpdateOrderStatus(ctx context.Context, order models.Order) error
}

// OrderService implements OrderService interface
type OrderService struct {
	repo    OrderRepository
	handler *accrual.Handler
}

// NewOrderService creates new OrderService instance
func NewOrderService(repo OrderRepository, handler *accrual.Handler) *OrderService {
	return &OrderService{repo: repo, handler: handler}
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

func (os *OrderService) doAccrualForOrder(order string) error {
	go func() {
		delay := 1 * time.Second
		var errTooManyReq *models.TooManyRequestsError

		for i := 0; i < 3; i++ {
			select {
			case <-time.After(delay):
				delay = 0
			}
			resp, err := os.handler.GetAccrualForOrder(context.TODO(), order)
			if err != nil {
				switch {
				case errors.As(err, errTooManyReq):
					delay = errTooManyReq.RetryAfter
					continue
				}
				return
			}

			curOrder, err := os.repo.GetOrderByNumber(context.TODO(), order)
			if err != nil {
				return
			}

			if strings.Compare(resp.Status, models.OrderStatusInvalid) == 0 && *resp.Accrual == 0 {
				return
			}
			// set new accrual and status
			curOrder.Accrual = resp.Accrual
			curOrder.Status = resp.Status

			os.repo.UpdateOrderStatus(context.TODO(), *curOrder)
		}
	}()

	return nil
}
