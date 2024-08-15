package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"
	"medods/internal/model"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	userRepo := NewUserRepository(db)

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
				mock.ExpectExec("insert into users").
					WithArgs(df.Email).
					WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectExec("insert into users").
					WithArgs(df.Email).
					WillReturnError(unexpectedError)
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
			err := userRepo.Create(context.Background(), test.input)
			test.checkResult(t, err)
		})
	}
}

func TestUserGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	userRepo := NewUserRepository(db)

	defaultUser := model.User{
		ID:    1,
		Email: "2",
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.User
		buildStubs  func()
		checkResult func(t *testing.T, in, db model.User, err error)
	}{
		{
			name:  "OK",
			input: defaultUser,
			buildStubs: func() {
				df := defaultUser
				mock.ExpectQuery("select id, email from users").
					WithArgs(df.ID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"email",
					}).AddRow(
						df.ID,
						df.Email,
					))
			},
			checkResult: func(t *testing.T, in, db model.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, in, db)
			},
		},
		{
			name:  "unexpected error",
			input: defaultUser,
			buildStubs: func() {
				df := defaultUser
				mock.ExpectQuery("select id, email from users").
					WithArgs(df.ID).
					WillReturnError(unexpectedError)
			},
			checkResult: func(t *testing.T, in, db model.User, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			user, err := userRepo.GetByID(context.Background(), test.input.ID)
			test.checkResult(t, test.input, user, err)
		})
	}
}

func TestUserList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	userRepo := NewUserRepository(db)

	defaultUser := model.User{
		ID:    1,
		Email: "2",
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.User
		buildStubs  func()
		checkResult func(t *testing.T, db []model.User, err error)
	}{
		{
			name:  "OK",
			input: defaultUser,
			buildStubs: func() {
				df1 := defaultUser
				df2 := defaultUser
				df2.ID = 2
				mock.ExpectQuery("select id, email from users").WithoutArgs().
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"email",
					}).AddRows(
						[]driver.Value{
							df1.ID,
							df1.Email,
						},
						[]driver.Value{
							df2.ID,
							df2.Email,
						},
					))
			},
			checkResult: func(t *testing.T, db []model.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 1, db[0].ID)
				assert.Equal(t, 2, db[1].ID)
			},
		},
		{
			name:  "unexpected error",
			input: defaultUser,
			buildStubs: func() {
				mock.ExpectQuery("select id, email from users").
					WithoutArgs().
					WillReturnError(unexpectedError)
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
			users, err := userRepo.List(context.Background())
			test.checkResult(t, users, err)
		})
	}
}
