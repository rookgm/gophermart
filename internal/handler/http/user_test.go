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

func TestUserHandler_RegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setup          func(t *testing.T) (*mocks.MockUserService, *mocks.MockTokenService)
		wantToken      string
		wantStatusCode int
	}{
		{
			// 200 — пользователь успешно зарегистрирован и аутентифицирован;
			name: "valid_request_return_200",
			body: `{"login": "user", "password": "secret"}`,
			setup: func(t *testing.T) (*mocks.MockUserService, *mocks.MockTokenService) {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				userMock := mocks.NewMockUserService(ctrl)
				userMock.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				tokenMock := mocks.NewMockTokenService(ctrl)
				tokenMock.EXPECT().CreateToken(gomock.Any()).Return("token", nil).AnyTimes()
				return userMock, tokenMock
			},
			wantToken:      `token`,
			wantStatusCode: http.StatusOK,
		},
		{
			// 400 — неверный формат запроса;
			name: "bad_request_return_400",
			body: "",
			setup: func(t *testing.T) (*mocks.MockUserService, *mocks.MockTokenService) {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				userMock := mocks.NewMockUserService(ctrl)
				userMock.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil).Times(0)

				tokenMock := mocks.NewMockTokenService(ctrl)
				tokenMock.EXPECT().CreateToken(gomock.Any()).Return("", nil).Times(0)
				return userMock, tokenMock
			},
			wantToken:      `token`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			// 409 — логин уже занят;
			name: "conflict_request_return_409",
			body: `{"login": "user", "password": "secret"}`,
			setup: func(t *testing.T) (*mocks.MockUserService, *mocks.MockTokenService) {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				userMock := mocks.NewMockUserService(ctrl)
				userMock.EXPECT().Register(gomock.Any(), gomock.Any()).Return(models.ErrConflictData).AnyTimes()

				tokenMock := mocks.NewMockTokenService(ctrl)
				tokenMock.EXPECT().CreateToken(gomock.Any()).Return("", nil).AnyTimes()
				return userMock, tokenMock
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			// 500 — внутренняя ошибка сервера.
			name: "internal_error_return_500",
			body: `{"login": "user", "password": "secret"}`,
			setup: func(t *testing.T) (*mocks.MockUserService, *mocks.MockTokenService) {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				userMock := mocks.NewMockUserService(ctrl)
				userMock.EXPECT().Register(gomock.Any(), gomock.Any()).Return(errors.New("internal error")).AnyTimes()

				tokenMock := mocks.NewMockTokenService(ctrl)
				tokenMock.EXPECT().CreateToken(gomock.Any()).Return("", nil).AnyTimes()
				return userMock, tokenMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}

			w := httptest.NewRecorder()
			svcUser, svcToken := tt.setup(t)

			handler := NewUserHandler(svcUser, svcToken)
			h := handler.RegisterUser()
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
