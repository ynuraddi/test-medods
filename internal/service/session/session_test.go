package session

import (
	"context"
	"fmt"
	"medods/internal/model"
	mock_repository "medods/internal/repository/mock"
	mock_logger "medods/pkg/logger/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSessionCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	sessRepo := mock_repository.NewMockSession(ctrl)

	service := New(sessRepo, logger)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		IP:         "5",
		CreatedAt:  6,
		Version:    7,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.Session
		buildStubs  func()
		checkResult func(t *testing.T, err error)
	}{
		{
			name:  "OK",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				sessRepo.EXPECT().Create(gomock.Any(), gomock.Eq(df)).Times(1).Return(nil)
			},
			checkResult: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "unexpected error",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				sessRepo.EXPECT().Create(gomock.Any(), gomock.Eq(df)).Times(1).Return(unexpectedError)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			err := service.Create(context.Background(), test.input)
			test.checkResult(t, err)
		})
	}
}

func TestSessionUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	sessRepo := mock_repository.NewMockSession(ctrl)

	service := New(sessRepo, logger)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		IP:         "5",
		CreatedAt:  6,
		Version:    7,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.Session
		buildStubs  func()
		checkResult func(t *testing.T, session model.Session, err error)
	}{
		{
			name:  "OK",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				sessRepo.EXPECT().Update(gomock.Any(), gomock.Eq(df)).Times(1).Return(defaultSession, nil)
			},
			checkResult: func(t *testing.T, db model.Session, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultSession, db)
			},
		},
		{
			name:  "unexpected error",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				sessRepo.EXPECT().Update(gomock.Any(), gomock.Eq(df)).Times(1).Return(model.Session{}, unexpectedError)
			},
			checkResult: func(t *testing.T, db model.Session, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
				assert.Equal(t, model.Session{}, db)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			session, err := service.Update(context.Background(), test.input)
			test.checkResult(t, session, err)
		})
	}
}

func TestUserGetByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	sessionRepo := mock_repository.NewMockSession(ctrl)

	service := New(sessionRepo, logger)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		IP:         "5",
		CreatedAt:  6,
		Version:    7,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.Session
		buildStubs  func()
		checkResult func(t *testing.T, db model.Session, err error)
	}{
		{
			name:  "OK",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				sessionRepo.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(df.UserID)).Times(1).Return(defaultSession, nil)
			},
			checkResult: func(t *testing.T, db model.Session, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultSession, db)
			},
		},
		{
			name:  "unexpected error",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				sessionRepo.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(df.UserID)).Times(1).Return(model.Session{}, unexpectedError)
			},
			checkResult: func(t *testing.T, db model.Session, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
				assert.Equal(t, model.Session{}, db)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			user, err := service.GetByUserID(context.Background(), test.input.UserID)
			test.checkResult(t, user, err)
		})
	}
}

func TestUserList(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	sessionRepo := mock_repository.NewMockSession(ctrl)

	service := New(sessionRepo, logger)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		IP:         "5",
		CreatedAt:  6,
		Version:    7,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		buildStubs  func()
		checkResult func(t *testing.T, db []model.Session, err error)
	}{
		{
			name: "OK",
			buildStubs: func() {
				df1 := defaultSession
				df2 := defaultSession
				df2.ID = 2
				sessionRepo.EXPECT().List(gomock.Any()).Times(1).Return([]model.Session{df1, df2}, nil)
			},
			checkResult: func(t *testing.T, db []model.Session, err error) {
				df1 := defaultSession
				df2 := defaultSession
				df2.ID = 2
				assert.NoError(t, err)
				assert.Equal(t, df1, db[0])
				assert.Equal(t, df2, db[1])
			},
		},
		{
			name: "unexpected error",
			buildStubs: func() {
				sessionRepo.EXPECT().List(gomock.Any()).Times(1).Return(nil, unexpectedError)
			},
			checkResult: func(t *testing.T, db []model.Session, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
				assert.Empty(t, db)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			users, err := service.List(context.Background())
			test.checkResult(t, users, err)
		})
	}
}
