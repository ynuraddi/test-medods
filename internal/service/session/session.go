package session

import (
	"context"
	"medods/internal/model"
	"medods/internal/repository"
	"medods/pkg/logger"
)

type Interface interface {
	Create(ctx context.Context, session model.Session) error
	Update(ctx context.Context, session model.Session) (model.Session, error)
	GetByUserID(ctx context.Context, id int) (model.Session, error)
	List(ctx context.Context) ([]model.Session, error)
}

var _ Interface = (*session)(nil)

type session struct {
	repo   repository.Session
	logger logger.Interface
}

func New(repo repository.Session, logger logger.Interface) *session {
	return &session{
		repo:   repo,
		logger: logger,
	}
}

func (s session) Create(ctx context.Context, session model.Session) error {
	return s.repo.Create(ctx, session)
}
func (s session) Update(ctx context.Context, session model.Session) (model.Session, error) {
	return s.repo.Update(ctx, session)
}
func (s session) GetByUserID(ctx context.Context, id int) (model.Session, error) {
	return s.repo.GetByUserID(ctx, id)
}
func (s session) List(ctx context.Context) ([]model.Session, error) {
	return s.repo.List(ctx)
}
