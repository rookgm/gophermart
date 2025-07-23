package service

import (
	"context"
	"errors"
	"github.com/phedde/luhn-algorithm"
	"github.com/rookgm/gophermart/internal/accrual"
	"github.com/rookgm/gophermart/internal/logger"
	"github.com/rookgm/gophermart/internal/models"
	"go.uber.org/zap"
	"strconv"
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
	// GetOrders returns orders with status NEW and PROCESSING
	GetOrders(ctx context.Context) ([]models.Order, error)
}

// OrderService implements OrderService interface
type OrderService struct {
	repo    OrderRepository
	handler *accrual.Handler
}

// NewOrderService creates new OrderService instance
func NewOrderService(repo OrderRepository, handler *accrual.Handler) *OrderService {
	return &OrderService{
		repo:    repo,
		handler: handler,
	}
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

// AccrualForOrder performs accrual for order
func (os *OrderService) AccrualForOrder(ctx context.Context, orderCh <-chan string) {
	for {
		var errTooManyReq models.TooManyRequestsError
		select {
		case <-ctx.Done():
			logger.Log.Debug("accrual is done")
			return
		case order, ok := <-orderCh:
			if !ok {
				return
			}

			logger.Log.Debug("try get accrual for order:", zap.String("number", order))
			resp, err := os.handler.GetAccrualForOrder(ctx, order)
			if err != nil {
				switch {
				case errors.As(err, &errTooManyReq):
					duration := errTooManyReq.RetryAfter
					logger.Log.Debug("too many request", zap.Duration("retry-after", duration))
					time.Sleep(duration)
					return
				}
				logger.Log.Error("accrual request error", zap.Error(err))
				return
			}

			logger.Log.Debug("accrual is ok, response:",
				zap.String("order", resp.Number),
				zap.String("status", resp.Status),
				zap.Float64p("accrual", resp.Accrual))

			curOrder, err := os.repo.GetOrderByNumber(ctx, order)
			if err != nil {
				logger.Log.Error("get order", zap.String("number", order))
				return
			}

			// set new accrual and status
			curOrder.Accrual = resp.Accrual
			curOrder.Status = resp.Status

			logger.Log.Debug("update order status", zap.String("number", order))
			if err := os.repo.UpdateOrderStatus(ctx, *curOrder); err != nil {
				logger.Log.Error("update order status", zap.String("number", order))
			}

			logger.Log.Debug("order status has been updated successfully", zap.String("number", order))
		}
	}
}

// GetOrdersForAccrual writes order to channel for accrual
func (os *OrderService) GetOrdersForAccrual(ctx context.Context, orderCh chan<- string) error {
	orders, err := os.repo.GetOrders(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		orderCh <- order.Number
	}

	return nil
}
