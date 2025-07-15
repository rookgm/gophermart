package handler

import (
	"context"
	"encoding/json"
	"errors"
	_ "github.com/jackc/pgx/v5/pgconn/ctxwatch"
	"github.com/rookgm/gophermart/internal/models"
	"net/http"
	"time"
)

type contextKey int

const (
	contextKeyUserID contextKey = iota
)

type BalanceService interface {
	// GetBalance returns current user balance
	GetBalance(ctx context.Context, userID uint64) (models.Balance, error)
	// BalanceWithdrawal withdrawals of balance
	BalanceWithdrawal(ctx context.Context, withdraw *models.Withdraw) (*models.Withdraw, error)
	// GetWithdrawals returns user withdrawals
	GetWithdrawals(ctx context.Context, userID uint64) ([]models.Withdraw, error)
}

// BalanceHandler represents HTTP handler for balance-related requests
type BalanceHandler struct {
	svc BalanceService
}

// NewBalanceHandler creates new BalanceHandler instance
func NewBalanceHandler(svc BalanceService) *BalanceHandler {
	return &BalanceHandler{svc: svc}
}

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// GetUserBalance returns current user balance
// 200 — успешная обработка запроса.
// 401 — пользователь не авторизован.
// 500 — внутренняя ошибка сервера.
func (bh *BalanceHandler) GetUserBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract user id
		userID, ok := r.Context().Value(contextKeyUserID).(uint64)
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		balance, err := bh.svc.GetBalance(r.Context(), userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		balanceResp := balanceResponse{
			Current:   balance.Current,
			Withdrawn: balance.Withdrawn,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(balanceResp); err != nil {
			return
		}
	}
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// UserBalanceWithdrawal performs user request for withdrawal
// 200 — успешная обработка запроса;
// 401 — пользователь не авторизован;
// 402 — на счету недостаточно средств;
// 422 — неверный номер заказа;
// 500 — внутренняя ошибка сервера.
func (bh *BalanceHandler) UserBalanceWithdrawal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract user id
		userID, ok := r.Context().Value(contextKeyUserID).(uint64)
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var withdrawReq withdrawRequest

		if err := json.NewDecoder(r.Body).Decode(&withdrawReq); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		withdraw := models.Withdraw{
			UserID:      userID,
			OrderNumber: withdrawReq.Order,
			Sum:         withdrawReq.Sum,
		}

		_, err := bh.svc.BalanceWithdrawal(r.Context(), &withdraw)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrInsufficientBalance):
				http.Error(w, "insufficient balance", http.StatusUnprocessableEntity)
			case errors.Is(err, models.ErrInvalidOrderNumber):
				http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
			case errors.Is(err, models.ErrOrderExist):
				http.Error(w, "order already exists", http.StatusUnprocessableEntity)
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type withdrawalsResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// GetUserWithdrawals gets withdrawal
// 200 — успешная обработка запроса.
// 204 — нет ни одного списания.
// 401 — пользователь не авторизован.
// 500 — внутренняя ошибка сервера.
func (bh *BalanceHandler) GetUserWithdrawals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract user id
		userID, ok := r.Context().Value(contextKeyUserID).(uint64)
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		withdrawals, err := bh.svc.GetWithdrawals(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrWithdrawalsNotExist):
				http.Error(w, "no content", http.StatusNoContent)
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		var resp []withdrawalsResponse

		for _, withdrawal := range withdrawals {
			resp = append(resp, withdrawalsResponse{
				Order:       withdrawal.OrderNumber,
				Sum:         withdrawal.Sum,
				ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			return
		}
	}
}
