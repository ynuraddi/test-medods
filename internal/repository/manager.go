package repository

import (
	"context"
	"database/sql"
	"medods/internal/model"
	"medods/internal/repository/postgres"
)

type Manager struct {
	User    User
	Session Session
}

func New(conn *sql.DB) *Manager {
	userRepo := postgres.NewUserRepository(conn)
	sessionRepo := postgres.NewSessionRepository(conn)

	return &Manager{
		User:    userRepo,
		Session: sessionRepo,
	}
}

type User interface {
	Create(ctx context.Context, u model.User) error
	GetByID(ctx context.Context, id int) (model.User, error)
	List(ctx context.Context) ([]model.User, error)
}

type Session interface {
	Create(ctx context.Context, session model.Session) error
	Update(ctx context.Context, session model.Session) (model.Session, error)
	GetByUserID(ctx context.Context, id int) (model.Session, error)
	List(ctx context.Context) ([]model.Session, error)
}
