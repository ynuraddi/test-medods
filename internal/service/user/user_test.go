package user

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

func TestUserCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	userRepo := mock_repository.NewMockUser(ctrl)

	service := New(userRepo, logger)

	defaultUser := model.User{
		ID:    1,
		Email: "2",
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.User
		buildStubs  func()
		checkResult func(t *testing.T, err error)
	}{
		{
			name:  "OK",
			input: defaultUser,
			buildStubs: func() {
				df := defaultUser
				userRepo.EXPECT().Create(gomock.Any(), gomock.Eq(df)).Times(1).Return(nil)
			},
			checkResult: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "unexpected error",
			input: defaultUser,
			buildStubs: func() {
				df := defaultUser
				userRepo.EXPECT().Create(gomock.Any(), gomock.Eq(df)).Times(1).Return(unexpectedError)
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

func TestUserGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	userRepo := mock_repository.NewMockUser(ctrl)

	service := New(userRepo, logger)

	defaultUser := model.User{
		ID:    1,
		Email: "2",
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.User
		buildStubs  func()
		checkResult func(t *testing.T, db model.User, err error)
	}{
		{
			name:  "OK",
			input: defaultUser,
			buildStubs: func() {
				df := defaultUser
				userRepo.EXPECT().GetByID(gomock.Any(), gomock.Eq(df.ID)).Times(1).Return(defaultUser, nil)
			},
			checkResult: func(t *testing.T, db model.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultUser, db)
			},
		},
		{
			name:  "unexpected error",
			input: defaultUser,
			buildStubs: func() {
				df := defaultUser
				userRepo.EXPECT().GetByID(gomock.Any(), gomock.Eq(df.ID)).Times(1).Return(model.User{}, unexpectedError)
			},
			checkResult: func(t *testing.T, db model.User, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
				assert.Equal(t, model.User{}, db)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			user, err := service.GetByID(context.Background(), test.input.ID)
			test.checkResult(t, user, err)
		})
	}
}

func TestUserList(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := mock_logger.NewMockInterface(ctrl)
	userRepo := mock_repository.NewMockUser(ctrl)

	service := New(userRepo, logger)

	defaultUser := model.User{
		ID:    1,
		Email: "2",
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		buildStubs  func()
		checkResult func(t *testing.T, db []model.User, err error)
	}{
		{
			name: "OK",
			buildStubs: func() {
				df1 := defaultUser
				df2 := defaultUser
				df2.ID = 2
				userRepo.EXPECT().List(gomock.Any()).Times(1).Return([]model.User{df1, df2}, nil)
			},
			checkResult: func(t *testing.T, db []model.User, err error) {
				df1 := defaultUser
				df2 := defaultUser
				df2.ID = 2
				assert.NoError(t, err)
				assert.Equal(t, df1, db[0])
				assert.Equal(t, df2, db[1])
			},
		},
		{
			name: "unexpected error",
			buildStubs: func() {
				userRepo.EXPECT().List(gomock.Any()).Times(1).Return(nil, unexpectedError)
			},
			checkResult: func(t *testing.T, db []model.User, err error) {
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
