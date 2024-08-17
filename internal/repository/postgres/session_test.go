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

func TestSessionCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	session := NewSessionRepository(db)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		CreatedAt:  5,
		Version:    6,
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
				mock.ExpectExec("insert into sessions").WithArgs(
					df.UserID,
					df.ATokenID,
					df.RTokenHash,
					df.CreatedAt,
				).WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectExec("insert into sessions").WithArgs(
					df.UserID,
					df.ATokenID,
					df.RTokenHash,
					df.CreatedAt,
				).WillReturnError(unexpectedError)
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
			err := session.Create(context.Background(), test.input)
			test.checkResult(t, err)
		})
	}
}

func TestSessionUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	session := NewSessionRepository(db)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		CreatedAt:  5,
		Version:    6,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.Session
		buildStubs  func()
		checkResult func(t *testing.T, in, db model.Session, err error)
	}{
		{
			name:  "OK",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				mock.ExpectQuery("update sessions").WithArgs(
					df.ID,
					df.Version,
					df.UserID,
					df.ATokenID,
					df.RTokenHash,
					df.CreatedAt,
				).WillReturnRows(sqlmock.NewRows([]string{
					"id",
					"user_id",
					"access_token_id",
					"refresh_token_hash",
					"created_at",
					"version",
				}).AddRow(
					df.ID,
					df.UserID,
					df.ATokenID,
					df.RTokenHash,
					df.CreatedAt,
					df.Version+1,
				))

			},
			checkResult: func(t *testing.T, in, db model.Session, err error) {
				assert.NoError(t, err)
				in.Version += 1
				assert.Equal(t, in, db)
			},
		},
		{
			name:  "unexpected error",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				mock.ExpectQuery("update sessions").WithArgs(
					df.ID,
					df.Version,
					df.UserID,
					df.ATokenID,
					df.RTokenHash,
					df.CreatedAt,
				).WillReturnError(unexpectedError)
			},
			checkResult: func(t *testing.T, in, db model.Session, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			session, err := session.Update(context.Background(), test.input)
			test.checkResult(t, test.input, session, err)
		})
	}
}

func TestSessionGetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	session := NewSessionRepository(db)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		CreatedAt:  5,
		Version:    6,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.Session
		buildStubs  func()
		checkResult func(t *testing.T, in, db model.Session, err error)
	}{
		{
			name:  "OK",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				mock.ExpectQuery("select id, user_id, access_token_id, refresh_token_hash, created_at, version from sessions").
					WithArgs(df.ID).
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"user_id",
						"access_token_id",
						"refresh_token_hash",
						"created_at",
						"version",
					}).AddRow(
						df.ID,
						df.UserID,
						df.ATokenID,
						df.RTokenHash,
						df.CreatedAt,
						df.Version,
					))

			},
			checkResult: func(t *testing.T, in, db model.Session, err error) {
				assert.NoError(t, err)
				assert.Equal(t, in, db)
			},
		},
		{
			name:  "unexpected error",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				mock.ExpectQuery(`select id, user_id, access_token_id, refresh_token_hash, created_at, version from sessions`).
					WithArgs(df.ID).
					WillReturnError(unexpectedError)
			},
			checkResult: func(t *testing.T, in, db model.Session, err error) {
				assert.Error(t, err)
				assert.Equal(t, err, unexpectedError)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			session, err := session.GetByUserID(context.Background(), test.input.ID)
			test.checkResult(t, test.input, session, err)
		})
	}
}

func TestSessionList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create db mock error: %s", err.Error())
	}

	session := NewSessionRepository(db)

	defaultSession := model.Session{
		ID:         1,
		UserID:     2,
		ATokenID:   "3",
		RTokenHash: "4",
		CreatedAt:  5,
		Version:    6,
	}

	unexpectedError := fmt.Errorf("unexpected error")

	tc := []struct {
		name        string
		input       model.Session
		buildStubs  func()
		checkResult func(t *testing.T, db []model.Session, err error)
	}{
		{
			name:  "OK",
			input: defaultSession,
			buildStubs: func() {
				df := defaultSession
				mock.ExpectQuery(`
					select 
						id, user_id, access_token_id, refresh_token_hash, created_at, version 
					from sessions`).
					WithoutArgs().
					WillReturnRows(sqlmock.NewRows([]string{
						"id",
						"user_id",
						"access_token_id",
						"refresh_token_hash",
						"created_at",
						"version",
					}).AddRows([]driver.Value{
						df.ID,
						df.UserID,
						df.ATokenID,
						df.RTokenHash,
						df.CreatedAt,
						df.Version,
					}, []driver.Value{
						df.ID + 1,
						df.UserID + 1,
						df.ATokenID,
						df.RTokenHash,
						df.CreatedAt,
						df.Version,
					}))

			},
			checkResult: func(t *testing.T, db []model.Session, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 1, db[0].ID)
				assert.Equal(t, 2, db[1].ID)
			},
		},
		{
			name:  "unexpected error",
			input: defaultSession,
			buildStubs: func() {
				mock.ExpectQuery(`
					select 
						id, user_id, access_token_id, refresh_token_hash, created_at, version 
					from sessions`).
					WithoutArgs().
					WillReturnError(unexpectedError)
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
			sessions, err := session.List(context.Background())
			test.checkResult(t, sessions, err)
		})
	}
}
