package worker

import (
	"context"
	"github.com/rookgm/gophermart/internal/logger"
	"time"
)

type OrderService interface {
	AccrualForOrder(ctx context.Context, orderCh <-chan string)
	GetOrdersForAccrual(ctx context.Context, orderCh chan<- string) error
}

// OrderProcessor is worker performs accrual for order
type OrderProcessor struct {
	svc OrderService
}

// NewOrderProcessor create new order processor
func NewOrderProcessor(svc OrderService) *OrderProcessor {
	return &OrderProcessor{svc: svc}
}

// ProcessOrders
func (op *OrderProcessor) ProcessOrders(ctx context.Context) {
	orders := make(chan string, 10)

	go op.svc.AccrualForOrder(ctx, orders)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("order processor is done")
			return
		case <-ticker.C:
			if err := op.svc.GetOrdersForAccrual(ctx, orders); err != nil {
				logger.Log.Error("error get order for accrual")
			}
		}
	}
}
