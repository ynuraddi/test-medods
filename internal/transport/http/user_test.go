package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"medods/internal/model"
	"medods/internal/service"
	mock_auth "medods/internal/service/auth/mock"
	mock_session "medods/internal/service/session/mock"
	mock_user "medods/internal/service/user/mock"
	"medods/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserCreate(t *testing.T) {
	ctrl := gomock.NewController(t)

	authService := mock_auth.NewMockInterface(ctrl)
	userService := mock_user.NewMockInterface(ctrl)
	sessionService := mock_session.NewMockInterface(ctrl)

	logger := logger.New("debug", true)

	router := NewRouter(&service.Manager{
		Auth:    authService,
		User:    userService,
		Session: sessionService,
	}, logger)

	defaultEmail := "mock@gmail.com"

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name          string
		input         string
		buildStubs    func()
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			input: defaultEmail,
			buildStubs: func() {
				userService.EXPECT().Create(gomock.Any(), gomock.Eq(model.User{Email: defaultEmail})).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name:       "error empty required field in request",
			input:      "",
			buildStubs: func() {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "error unexpected user create",
			input: defaultEmail,
			buildStubs: func() {
				userService.EXPECT().Create(gomock.Any(), gomock.Eq(model.User{Email: defaultEmail})).Times(1).Return(unexpectedError)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()

			j, err := json.Marshal(createUserRequest{test.input})
			assert.NoError(t, err)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user/create", bytes.NewBuffer(j))

			router.ServeHTTP(rec, req)

			test.checkResponse(t, rec)
		})
	}
}

func TestUserList(t *testing.T) {
	ctrl := gomock.NewController(t)

	authService := mock_auth.NewMockInterface(ctrl)
	userService := mock_user.NewMockInterface(ctrl)
	sessionService := mock_session.NewMockInterface(ctrl)

	logger := logger.New("debug", true)

	router := NewRouter(&service.Manager{
		Auth:    authService,
		User:    userService,
		Session: sessionService,
	}, logger)

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name          string
		buildStubs    func()
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			buildStubs: func() {
				userService.EXPECT().List(gomock.Any()).Times(1).Return([]model.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK no content",
			buildStubs: func() {
				userService.EXPECT().List(gomock.Any()).Times(1).Return([]model.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name: "error unexpected session list",
			buildStubs: func() {
				userService.EXPECT().List(gomock.Any()).Times(1).Return(nil, unexpectedError)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/user/list", nil)

			router.ServeHTTP(rec, req)

			test.checkResponse(t, rec)
		})
	}
}
