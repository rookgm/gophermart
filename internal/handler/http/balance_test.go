package handler

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/rookgm/gophermart/internal/handler/http/mocks"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBalanceHandler_GetUserBalance(t *testing.T) {
	tests := []struct {
		name           string
		token          *models.TokenPayload
		setup          func(t *testing.T) *mocks.MockBalanceService
		wantStatusCode int
		wantBody       *balanceResponse
	}{
		{
			// 200 — успешная обработка запроса.
			name: "valid_request_return_200",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().GetBalance(gomock.Any(), gomock.Any()).Return(models.Balance{
					Current:   100,
					Withdrawn: 100,
				}, nil).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusOK,
			wantBody: &balanceResponse{
				Current:   0,
				Withdrawn: 100,
			},
		},
		{
			// 500 — внутренняя ошибка сервера.
			name: "internal_error_return_500",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().GetBalance(gomock.Any(), gomock.Any()).Return(models.Balance{}, models.ErrInternalError).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/user/balance", nil)
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			st := tt.setup(t)
			ctx := context.WithValue(req.Context(), authPayloadKey, tt.token)

			handler := NewBalanceHandler(st)
			h := handler.GetUserBalance()
			h(w, req.WithContext(ctx))

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.wantBody != nil {
				var got balanceResponse
				err = json.Unmarshal(resBody, &got)
				require.NoError(t, err)

				if diff := cmp.Diff(*tt.wantBody, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestBalanceHandler_UserBalanceWithdrawal(t *testing.T) {
	tests := []struct {
		name           string
		token          *models.TokenPayload
		body           string
		setup          func(t *testing.T) *mocks.MockBalanceService
		wantStatusCode int
	}{
		{
			// 200 — успешная обработка запроса.
			name: "valid_request_return_200",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: `{"order": "2377225624", "sum": 751}`,
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().BalanceWithdrawal(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusOK,
		},
		{
			// 402 — на счету недостаточно средств;
			name: "insufficient_balance_return_402",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: `{"order": "2377225624", "sum": 751}`,
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().BalanceWithdrawal(gomock.Any(), gomock.Any()).Return(nil, models.ErrInsufficientBalance).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusPaymentRequired,
		},
		{
			// 422 — неверный номер заказа;
			name: "bad_order_number_return_422",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: `{"order": "1", "sum": 751}`,
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().BalanceWithdrawal(gomock.Any(), gomock.Any()).Return(nil, models.ErrInvalidOrderNumber).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			// 500 — внутренняя ошибка сервера.
			name: "internal_error_return_500",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: `{"order": "2377225624", "sum": 751}`,
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().BalanceWithdrawal(gomock.Any(), gomock.Any()).Return(nil, models.ErrInternalError).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			st := tt.setup(t)
			ctx := context.WithValue(req.Context(), authPayloadKey, tt.token)

			handler := NewBalanceHandler(st)
			h := handler.UserBalanceWithdrawal()
			h(w, req.WithContext(ctx))

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			defer res.Body.Close()

		})
	}
}

func TestBalanceHandler_GetUserWithdrawals(t *testing.T) {
	processedAt := time.Now()
	tests := []struct {
		name           string
		token          *models.TokenPayload
		setup          func(t *testing.T) *mocks.MockBalanceService
		wantStatusCode int
		wantBody       []withdrawalsResponse
	}{
		{
			// 200 — успешная обработка запроса.
			name: "valid_request_return_200",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Return([]models.Withdraw{
					{
						ID:          1,
						UserID:      1,
						OrderNumber: "2377225624",
						Sum:         500,
						ProcessedAt: processedAt,
					},
				}, nil).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusOK,
			wantBody: []withdrawalsResponse{{
				Order:       "2377225624",
				Sum:         500,
				ProcessedAt: processedAt.Format(time.RFC3339),
			}},
		},
		{
			// 204 — нет ни одного списания.
			name: "no_content_return_204",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Return(nil, models.ErrWithdrawalsNotExist).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusNoContent,
			wantBody:       nil,
		},
		{
			// 500 — внутренняя ошибка сервера.
			name: "internal_error_return_500",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockBalanceService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockBalanceService(ctrl)
				svcMock.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Return(nil, models.ErrInternalError).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			st := tt.setup(t)
			ctx := context.WithValue(req.Context(), authPayloadKey, tt.token)

			handler := NewBalanceHandler(st)
			h := handler.GetUserWithdrawals()
			h(w, req.WithContext(ctx))

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.wantBody != nil {
				var got []withdrawalsResponse
				err = json.Unmarshal(resBody, &got)
				require.NoError(t, err)

				if diff := cmp.Diff(tt.wantBody, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
