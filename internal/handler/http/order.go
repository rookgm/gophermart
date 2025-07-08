package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rookgm/gophermart/internal/models"
	"io"
	"net/http"
	"time"
)

type OrderService interface {
	// Upload uploads user order
	Upload(ctx context.Context, order *models.Order) (*models.Order, error)
	// ListUserOrders returns list of user orders
	ListUserOrders(ctx context.Context, userID uint64) ([]models.Order, error)
}

// OrderHandler represents HTTP handler for order-related requests
type OrderHandler struct {
	svc OrderService
}

// NewOrderHandler creates new OrderService instance
// 200 — номер заказа уже был загружен этим пользователем;
// 202 — новый номер заказа принят в обработку;
// 400 — неверный формат запроса;
// 401 — пользователь не аутентифицирован;
// 409 — номер заказа уже был загружен другим пользователем;
// 422 — неверный формат номера заказа;
// 500 — внутренняя ошибка сервера.
func NewOrderHandler(svc OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// UploadOrder uploads user order
func (oh *OrderHandler) UploadOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract user id
		userID, ok := r.Context().Value("userid").(uint64)
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// get order id
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ord := models.Order{
			UserID: userID,
			Number: string(body),
		}

		_, err = oh.svc.Upload(r.Context(), &ord)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrInvalidOrderID):
				http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
			case errors.Is(err, models.ErrOrderLoadedUser):
				http.Error(w, "order has already been uploaded", http.StatusOK)
			case errors.Is(err, models.ErrOrderLoadedAnotherUser):
				http.Error(w, "order has already been uploaded by another user", http.StatusUnprocessableEntity)
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

type ListOrdersResp struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

// ListOrders get list uploaded user orders
// 200 — успешная обработка запроса.
// 204 — нет данных для ответа.
// 401 — пользователь не авторизован.
// 500 — внутренняя ошибка сервера.
func (oh *OrderHandler) ListOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract user id
		userID, ok := r.Context().Value("userid").(uint64)
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// get user orders
		orders, err := oh.svc.ListUserOrders(r.Context(), userID)
		if err != nil {
			if errors.Is(err, models.ErrDataNotFound) {
				http.Error(w, "orders not found", http.StatusNoContent)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var ordersResp []ListOrdersResp

		for _, order := range orders {
			ordersResp = append(ordersResp, ListOrdersResp{
				Number:     order.Number,
				Status:     order.Status,
				Accrual:    order.Accrual,
				UploadedAt: order.UploadedAt.Format(time.RFC3339),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(ordersResp); err != nil {
			return
		}
	}
}
