package handler

import (
	"context"
	"encoding/json"
	"errors"
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

func TestOrderHandler_UploadUserOrder(t *testing.T) {
	tests := []struct {
		name           string
		token          *models.TokenPayload
		body           string
		setup          func(t *testing.T) *mocks.MockOrderService
		wantStatusCode int
	}{
		{
			// 202 — новый номер заказа принят в обработку;
			name: "valid_request_return_202",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: "286436514",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusAccepted,
		},
		{
			// 200 — номер заказа уже был загружен этим пользователем;
			name: "valid_request_return_200",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: "286436514",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, models.ErrOrderLoadedUser).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusOK,
		},
		{
			// 400 — неверный формат запроса(пустой номер заказа);
			name: "bad_request_return_400",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: "",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, errors.New("bad request")).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			// 401 — пользователь не аутентифицирован;
			name: "unauthorized_request_return_401",
			body: "286436514",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, errors.New("unauthorized")).Times(0)
				return svcMock
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			// 409 — номер заказа уже был загружен другим пользователем;
			name: "conflict_request_return_409",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: "286436514",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, models.ErrOrderLoadedAnotherUser).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			// 422 — неверный формат номера заказа;
			name: "invalid_order_number_request_return_409",
			token: &models.TokenPayload{
				UserID: 1,
			},
			body: "1",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, models.ErrInvalidOrderNumber).AnyTimes()
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
			body: "286436514",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(nil, models.ErrInternalError).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			st := tt.setup(t)
			ctx := context.WithValue(req.Context(), authPayloadKey, tt.token)

			handler := NewOrderHandler(st)
			h := handler.UploadUserOrder()
			h(w, req.WithContext(ctx))

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			defer res.Body.Close()
		})
	}
}

func TestOrderHandler_ListUserOrders(t *testing.T) {
	uploadAt := time.Now()
	tests := []struct {
		name           string
		token          *models.TokenPayload
		setup          func(t *testing.T) *mocks.MockOrderService
		wantStatusCode int
		wantBody       []ListOrdersResp
	}{
		{
			// 200 — успешная обработка запроса.
			name: "valid_request_return_200",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().ListUserOrders(gomock.Any(), gomock.Any()).Return([]models.Order{
					{
						ID:         1,
						UserID:     1,
						Number:     "12345678903",
						Status:     "PROCESSED",
						UploadedAt: uploadAt,
					},
				}, nil).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusOK,
			wantBody: []ListOrdersResp{{
				Number:     "12345678903",
				Status:     "PROCESSED",
				UploadedAt: uploadAt.Format(time.RFC3339),
			}},
		},
		{
			// 204 — нет данных для ответа.
			name: "not_content_request_return_204",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().ListUserOrders(gomock.Any(), gomock.Any()).Return(nil, models.ErrDataNotFound).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			// 401 — пользователь не авторизован.
			name: "unauthorized_request_return_401",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().ListUserOrders(gomock.Any(), gomock.Any()).Return(nil, models.ErrDataNotFound).Times(0)
				return svcMock
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			// 500 — внутренняя ошибка сервера.
			name: "internal_error_return_500",
			token: &models.TokenPayload{
				UserID: 1,
			},
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().ListUserOrders(gomock.Any(), gomock.Any()).Return(nil, models.ErrInternalError).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			st := tt.setup(t)
			ctx := context.WithValue(req.Context(), authPayloadKey, tt.token)

			handler := NewOrderHandler(st)
			h := handler.ListUserOrders()
			h(w, req.WithContext(ctx))

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.wantBody != nil {
				var got []ListOrdersResp
				err = json.Unmarshal(resBody, &got)
				require.NoError(t, err)

				if diff := cmp.Diff(tt.wantBody, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
