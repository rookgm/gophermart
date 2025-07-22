package handler

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/rookgm/gophermart/internal/handler/http/mocks"
	"github.com/rookgm/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthHandler_LoginUser(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setup          func(t *testing.T) *mocks.MockAuthService
		wantToken      string
		wantStatusCode int
	}{
		{
			// 200 — пользователь успешно аутентифицирован;
			name: "valid_request_return_200",
			body: `{"login": "user", "password": "secret"}`,
			setup: func(t *testing.T) *mocks.MockAuthService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				authMock := mocks.NewMockAuthService(ctrl)
				authMock.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return("token", nil).AnyTimes()

				return authMock
			},
			wantToken:      `token`,
			wantStatusCode: http.StatusOK,
		},
		{
			// 400 — неверный формат запроса;
			name: "bad_request_return_400",
			body: "",
			setup: func(t *testing.T) *mocks.MockAuthService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				authMock := mocks.NewMockAuthService(ctrl)
				authMock.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return("token", nil).Times(0)

				return authMock
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			// 401 — неверная пара логин/пароль;
			name: "unauthorized_request_return_401",
			body: `{"login": "user", "password": "secret"}`,
			setup: func(t *testing.T) *mocks.MockAuthService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				authMock := mocks.NewMockAuthService(ctrl)
				authMock.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return("", models.ErrInvalidCredentials).AnyTimes()

				return authMock
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			// 500 — внутренняя ошибка сервера.
			name: "internal_error_return_500",
			body: `{"login": "user", "password": "secret"}`,
			setup: func(t *testing.T) *mocks.MockAuthService {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				authMock := mocks.NewMockAuthService(ctrl)
				authMock.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("internal error")).AnyTimes()

				return authMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			st := tt.setup(t)

			handler := NewAuthHandler(st)
			h := handler.LoginUser()
			h(w, req)

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			defer res.Body.Close()

			for _, cookie := range res.Cookies() {
				if cookie.Name == "auth_token" {
					assert.Equal(t, tt.wantToken, cookie.Value)
					break
				}
			}
		})
	}
}
