package http

import (
	"database/sql"
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

func TestSessionList(t *testing.T) {
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
				sessionService.EXPECT().List(gomock.Any()).Times(1).Return([]model.Session{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK no content",
			buildStubs: func() {
				sessionService.EXPECT().List(gomock.Any()).Times(1).Return([]model.Session{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name: "error unexpected session list",
			buildStubs: func() {
				sessionService.EXPECT().List(gomock.Any()).Times(1).Return(nil, unexpectedError)
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
			req := httptest.NewRequest(http.MethodGet, "/api/v1/session/list", nil)

			router.ServeHTTP(rec, req)

			test.checkResponse(t, rec)
		})
	}
}
