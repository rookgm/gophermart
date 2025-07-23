package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/rookgm/gophermart/internal/repository/postgres"
)

const (
	insertOrderQuery = `
						INSERT INTO orders (user_id, number, status) 
						values ($1, $2, $3)
						RETURNING id, user_id, number, status, accrual, uploaded_at
`
	selectOrderByNumQuery = `
						SELECT id, user_id, number, status, accrual, uploaded_at FROM orders
						WHERE number = $1
`

	selectOrdersByUserIDQuery = `
						SELECT id, user_id, number, status, accrual, uploaded_at FROM orders
						WHERE user_id = $1
						ORDER BY uploaded_at DESC
`
	updateOrderQuery = `
						UPDATE orders
						SET status = $1, accrual = $2
						WHERE number = $3
`

	selectOrders = `
						SELECT id, user_id, number, status, accrual, uploaded_at FROM orders
						WHERE status IN ('NEW', 'PROCESSING')
	`
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
	err := or.db.QueryRow(ctx, insertOrderQuery, order.UserID, order.Number, order.Status).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errCode := or.db.ErrorCode(err); errCode == pgErrUniqueViolationCode {
			return nil, models.ErrConflictData
		}
		return nil, err
	}

	return order, nil
}

// GetOrderByNumber returns order by number
func (or *OrderRepository) GetOrderByNumber(ctx context.Context, num string) (*models.Order, error) {
	order := models.Order{}
	err := or.db.QueryRow(ctx, selectOrderByNumQuery, num).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrDataNotFound
		}
		return nil, err
	}

	return &order, nil
}

// GetOrdersByUserID gets user orders
func (or *OrderRepository) GetOrdersByUserID(ctx context.Context, userID uint64) ([]models.Order, error) {
	rows, err := or.db.Query(ctx, selectOrdersByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.Order{}

	for rows.Next() {
		order := models.Order{}
		err = rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// UpdateOrderStatus update order status and accrual
func (or *OrderRepository) UpdateOrderStatus(ctx context.Context, order models.Order) error {
	cmd, err := or.db.Exec(ctx, updateOrderQuery, order.Status, order.Accrual, order.Number)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return models.ErrDataNotFound
	}

	return nil
}

// GetOrders returns orders with status NEW and PROCESSING
func (or *OrderRepository) GetOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := or.db.Query(ctx, selectOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.Order{}

	for rows.Next() {
		order := models.Order{}
		err = rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
