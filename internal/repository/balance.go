package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/rookgm/gophermart/internal/repository/postgres"
)

const (
	insertWithdrawalQuery = `
						INSERT INTO withdrawals (user_id, order_number, amount) 
						values ($1, $2, $3)
						RETURNING id, user_id, order_number, amount, processed_at
`
	selectWithdrawalsByUserIDQuery = `
						SELECT id, user_id, order_number, amount, processed_at FROM withdrawals
						WHERE user_id = $1
						ORDER BY processed_at DESC
`
	selectBalanceQuery = `
						select 
    					(select coalesce(sum(accrual),0) 
    					 from orders 
    					 where user_id=$1 AND status='Processed'), 
						(select coalesce(sum(amount),0)
						 from withdrawals
						 where user_id=$1)
`
)

// BalanceRepository implements BalanceRepository interface
type BalanceRepository struct {
	db *postgres.DB
}

// NewBalanceRepository creates new balance repository instance
func NewBalanceRepository(db *postgres.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// Balance returns current balance
func (br *BalanceRepository) Balance(ctx context.Context, userID uint64) (models.Balance, error) {
	balance := models.Balance{}
	err := br.db.QueryRow(ctx, selectBalanceQuery, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Balance{}, models.ErrDataNotFound
		}
		return models.Balance{}, err
	}
	return balance, nil
}

// CreateWithdrawal creates new withdrawal
func (br *BalanceRepository) CreateWithdrawal(ctx context.Context, withdraw *models.Withdraw) (*models.Withdraw, error) {
	err := br.db.QueryRow(ctx, insertWithdrawalQuery, withdraw.UserID, withdraw.OrderNumber, withdraw.Sum).Scan(&withdraw.ID, &withdraw.UserID, &withdraw.OrderNumber, &withdraw.Sum, &withdraw.ProcessedAt)
	if err != nil {
		if errCode := br.db.ErrorCode(err); errCode == "23505" {
			return nil, models.ErrConflictData
		}
		return nil, err
	}

	return withdraw, nil
}

// GetWithdrawalsByUserID returns withdrawals
func (br *BalanceRepository) GetWithdrawalsByUserID(ctx context.Context, userID uint64) ([]models.Withdraw, error) {
	rows, err := br.db.Query(ctx, selectWithdrawalsByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdraws []models.Withdraw

	for rows.Next() {
		withdraw := models.Withdraw{}
		err = rows.Scan(&withdraw.ID, &withdraw.UserID, &withdraw.OrderNumber, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			continue
		}
		withdraws = append(withdraws, withdraw)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdraws, nil
}
