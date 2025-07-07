package handler

import (
	"context"
	"errors"
	"github.com/rookgm/gophermart/internal/models"
	"io"
	"net/http"
)

type OrderService interface {
	Upload(ctx context.Context, order *models.Order) (*models.Order, error)
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
			UserID:  userID,
			OrderID: string(body),
		}

		_, err = oh.svc.Upload(r.Context(), &ord)
		if err != nil {
			if errors.Is(err, models.ErrInvalidOrderID) {
				http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
				return
			}
		}
	}
}
