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
}

// OrderService implements OrderService interface
type OrderService struct {
	repo    OrderRepository
	handler *accrual.Handler
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewOrderService creates new OrderService instance
func NewOrderService(repo OrderRepository, handler *accrual.Handler) *OrderService {
	ctx, cancel := context.WithCancel(context.Background())
	return &OrderService{repo: repo,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
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

	os.doAccrualForOrder(order.Number)

	return order, nil
}

// ListUserOrders returns list of user orders
func (os *OrderService) ListUserOrders(ctx context.Context, userID uint64) ([]models.Order, error) {
	return os.repo.GetOrdersByUserID(ctx, userID)
}

// doAccrualForOrder performs accrual for order
func (os *OrderService) doAccrualForOrder(order string) error {
	go func() {
		var errTooManyReq models.TooManyRequestsError
		delay := time.Duration(0)
		logger.Log.Debug("starting accrual")

		for i := 1; i <= 3; i++ {
			t := time.NewTimer(delay)
			select {
			case <-os.ctx.Done():
				logger.Log.Debug("accrual is done")
				t.Stop()
				return
			case <-t.C:
				logger.Log.Debug("timeout")
			}
			logger.Log.Debug("attempt:", zap.Int("number", i))
			logger.Log.Debug("get accrual for order:", zap.String("number", order))
			resp, err := os.handler.GetAccrualForOrder(os.ctx, order)
			if err != nil {
				switch {
				case errors.As(err, &errTooManyReq):
					logger.Log.Debug("too many request")
					delay = errTooManyReq.RetryAfter
					continue
				}
				return
			}

			logger.Log.Debug("accrual response", zap.Any("order", resp))

			curOrder, err := os.repo.GetOrderByNumber(os.ctx, order)
			if err != nil {
				logger.Log.Error("get order", zap.String("number", order))
				return
			}

			// set new accrual and status
			curOrder.Accrual = resp.Accrual
			curOrder.Status = resp.Status

			logger.Log.Debug("update order status", zap.String("number", order))
			if err := os.repo.UpdateOrderStatus(os.ctx, *curOrder); err != nil {
				logger.Log.Error("update order status", zap.String("number", order))
			}

			logger.Log.Debug("order status has been updated successfully", zap.String("number", order))

			t.Stop()
			break
		}
	}()

	return nil
}

// StopAccrual stops accrual
func (os *OrderService) StopAccrual(ctx context.Context) {
	os.cancel()
	select {
	case <-os.ctx.Done():
		logger.Log.Info("accrual is canceled")
	case <-ctx.Done():
		logger.Log.Info("accrual is stopped")
	}
}
