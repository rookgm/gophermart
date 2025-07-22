package handler

import (
	"github.com/golang/mock/gomock"
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
		body           string
		setup          func(t *testing.T) *mocks.MockOrderService
		wantStatusCode int
	}{
		{
			name: "valid_request_return_202",
			body: "286436514",
			setup: func(t *testing.T) *mocks.MockOrderService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				svcMock := mocks.NewMockOrderService(ctrl)
				svcMock.EXPECT().Upload(gomock.Any(), gomock.Any()).Return(&models.Order{
					ID:         1,
					UserID:     1,
					Number:     "286436514",
					Status:     "NEW",
					Accrual:    nil,
					UploadedAt: time.Time{},
				}, nil).AnyTimes()
				return svcMock
			},
			wantStatusCode: http.StatusAccepted,
		},
		/*{
			name: "no_content_return_204",
			body: "",
			setup: func(t *testing.T) *storage.MockURLStorage {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				storeMock := storage.NewMockURLStorage(ctrl)
				storeMock.EXPECT().GetUserURLsCtx(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return storeMock
			},
			wantStatusCode: http.StatusNoContent,
			wantBody:       nil,
		},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()

			st := tt.setup(t)

			handler := NewOrderHandler(st)
			h := handler.UploadUserOrder()
			h(w, req)

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			defer res.Body.Close()

			_, err = io.ReadAll(res.Body)
			require.NoError(t, err)

		})
	}
}
