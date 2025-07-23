package accrual

import (
	"context"
	"encoding/json"
	"github.com/rookgm/gophermart/internal/models"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// default time of retry after
const delaySeconds = 60

// Handler represents HTTP handler for accrual-related requests
type Handler struct {
	client  *http.Client
	baseURL string
}

// NewAccrualClient creates new Handler instance
func NewAccrualClient(baseURL string) *Handler {
	return &Handler{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: baseURL,
	}
}

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// GetAccrualForOrder returns accrual for order
// 200 — успешная обработка запроса.
// 204 — заказ не зарегистрирован в системе расчёта.
// 429 — превышено количество запросов к сервису.
// 500 — внутренняя ошибка сервера.
func (h *Handler) GetAccrualForOrder(ctx context.Context, order string) (*models.Order, error) {
	// GET /api/orders/{number}
	url, err := url.JoinPath(h.baseURL, "api", "orders", order)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		acrResp := accrualResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&acrResp); err != nil {
			return nil, err
		}
		return &models.Order{
			Number:  acrResp.Order,
			Status:  acrResp.Status,
			Accrual: &acrResp.Accrual,
		}, nil
	case http.StatusNoContent:
		return nil, models.ErrOrderNotRegInAccrual
	case http.StatusTooManyRequests:
		var t int
		val := resp.Header.Get("Retry-After")
		if val == "" {
			// set default
			t = delaySeconds
		}
		t, err := strconv.Atoi(val)
		if err != nil {
			t = delaySeconds
		}
		return nil, models.NewTooManyRequestsError(time.Duration(t) * time.Second)
	case http.StatusInternalServerError:
		return nil, models.ErrInternalError
	default:
		return nil, err
	}
}
