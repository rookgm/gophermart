package service

import (
	"context"
	"errors"
	"github.com/phedde/luhn-algorithm"
	"github.com/rookgm/gophermart/internal/models"
	"strconv"
)

// BalanceRepository is interface for interfacing with balance-related data
type BalanceRepository interface {
	// Balance returns current balance
	Balance(ctx context.Context, userID uint64) (models.Balance, error)
	// CreateWithdrawal creates new withdrawal
	CreateWithdrawal(ctx context.Context, withdraw *models.Withdraw) (*models.Withdraw, error)
	// GetWithdrawalsByUserID returns withdrawals
	GetWithdrawalsByUserID(ctx context.Context, userID uint64) ([]models.Withdraw, error)
}

// BalanceService implements BalanceService interface
type BalanceService struct {
	repo BalanceRepository
}

// NewBalanceService creates new BalanceService instance
func NewBalanceService(repo BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

// GetBalance returns current user balance
func (bs *BalanceService) GetBalance(ctx context.Context, userID uint64) (models.Balance, error) {
	return bs.repo.Balance(ctx, userID)
}

// BalanceWithdrawal withdrawals of balance
func (bs *BalanceService) BalanceWithdrawal(ctx context.Context, withdraw *models.Withdraw) (*models.Withdraw, error) {
	num, err := strconv.ParseInt(withdraw.OrderNumber, 10, 64)
	if err != nil {
		return nil, err
	}

	// check order id using Luhn algorithm
	if ok := luhn.IsValid(num); !ok {
		return nil, models.ErrInvalidOrderNumber
	}

	// check insufficient balance
	balance, err := bs.GetBalance(ctx, withdraw.UserID)
	if err != nil {
		return nil, err
	}

	if withdraw.Sum > balance.Current-balance.Withdrawn {
		return nil, models.ErrInsufficientBalance
	}

	w, err := bs.repo.CreateWithdrawal(ctx, withdraw)
	if err != nil {
		if errors.Is(err, models.ErrConflictData) {
			return nil, models.ErrOrderExist
		}
	}

	return w, nil
}

// GetWithdrawals returns user withdrawals
func (bs *BalanceService) GetWithdrawals(ctx context.Context, userID uint64) ([]models.Withdraw, error) {
	withdrawls, err := bs.repo.GetWithdrawalsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(withdrawls) == 0 {
		return nil, models.ErrWithdrawalsNotExist
	}

	return withdrawls, nil
}
