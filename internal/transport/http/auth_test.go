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

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthLogin(t *testing.T) {
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

	type args struct {
		path string
		ip   string
	}

	defaultArgs := args{
		path: "1",
		ip:   "0.0.0.0",
	}

	defaultAToken := "access_token"
	defaultRToken := "rand_string"

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name          string
		input         args
		buildStubs    func()
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			input: defaultArgs,
			buildStubs: func() {
				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(1)).Times(1).Return(model.User{ID: 1, Email: "test"}, nil)
				authService.EXPECT().CreateSession(gomock.Any(), gomock.Eq(1), gomock.Eq(defaultArgs.ip)).Times(1).Return(defaultAToken, defaultRToken, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "error incorrect param",
			input: args{
				path: "incorrect", // note: incorrect
				ip:   defaultArgs.ip,
			},
			buildStubs: func() {}, // none expecting calls
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "error unexpected get user by id",
			input: defaultArgs,
			buildStubs: func() {
				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(1)).Times(1).Return(model.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:  "error unexpected get user by id",
			input: defaultArgs,
			buildStubs: func() {
				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(1)).Times(1).Return(model.User{}, unexpectedError)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "error unexpected create session",
			input: defaultArgs,
			buildStubs: func() {
				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(1)).Times(1).Return(model.User{ID: 1, Email: "test"}, nil)
				authService.EXPECT().CreateSession(gomock.Any(), gomock.Eq(1), gomock.Eq(defaultArgs.ip)).Times(1).Return("", "", unexpectedError)
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/auth/login/%s", test.input.path), nil)
			req.Header.Set("x-forwarded-for", test.input.ip)

			router.ServeHTTP(rec, req)

			test.checkResponse(t, rec)
		})
	}
}

func TestAuthRefresh(t *testing.T) {
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

	type args struct {
		aToken string
		rToken string
	}

	defaultAToken := "access_token"
	defaultRToken := "rand_string"

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name          string
		input         args
		buildStubs    func()
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: args{
				aToken: defaultAToken,
				rToken: defaultRToken,
			},
			buildStubs: func() {
				authService.EXPECT().RefreshSession(gomock.Any(), gomock.Eq(defaultAToken), gomock.Eq(defaultRToken)).Times(1).Return(defaultAToken, defaultRToken, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				j, err := json.Marshal(refreshResponse{
					AccessToken:  defaultAToken,
					RefreshToken: defaultRToken,
				})
				assert.NoError(t, err)

				assert.Equal(t, http.StatusOK, recorder.Code)
				assert.Equal(t, string(j), recorder.Body.String())
			},
		},
		{
			name: "error no access token in header",
			input: args{
				aToken: "",
				rToken: defaultRToken,
			},
			buildStubs: func() {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				j, err := json.Marshal(map[string]any{
					"error": "authorization token is empty",
				})
				assert.NoError(t, err)

				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assert.Equal(t, string(j), recorder.Body.String())
			},
		},
		{
			name: "error no refresh token in body",
			input: args{
				aToken: defaultAToken,
				rToken: "",
			},
			buildStubs: func() {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "error refresh session token expired",
			input: args{
				aToken: defaultAToken,
				rToken: defaultRToken,
			},
			buildStubs: func() {
				authService.EXPECT().RefreshSession(gomock.Any(), gomock.Eq(defaultAToken), gomock.Eq(defaultRToken)).Times(1).Return("", "", jwt.ErrTokenExpired)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "error refresh session token invalid signature",
			input: args{
				aToken: defaultAToken,
				rToken: defaultRToken,
			},
			buildStubs: func() {
				authService.EXPECT().RefreshSession(gomock.Any(), gomock.Eq(defaultAToken), gomock.Eq(defaultRToken)).Times(1).Return("", "", jwt.ErrSignatureInvalid)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "error refresh session token malfored",
			input: args{
				aToken: defaultAToken,
				rToken: defaultRToken,
			},
			buildStubs: func() {
				authService.EXPECT().RefreshSession(gomock.Any(), gomock.Eq(defaultAToken), gomock.Eq(defaultRToken)).Times(1).Return("", "", jwt.ErrTokenMalformed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "error unexpected refresh session",
			input: args{
				aToken: defaultAToken,
				rToken: defaultRToken,
			},
			buildStubs: func() {
				authService.EXPECT().RefreshSession(gomock.Any(), gomock.Eq(defaultAToken), gomock.Eq(defaultRToken)).Times(1).Return("", "", unexpectedError)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()

			j, err := json.Marshal(refreshRequest{test.input.rToken})
			assert.NoError(t, err)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(j))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.input.aToken))

			router.ServeHTTP(rec, req)

			test.checkResponse(t, rec)
		})
	}
}
