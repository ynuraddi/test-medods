package user

import (
	"context"
	"medods/internal/model"
	"medods/internal/repository"
	"medods/pkg/logger"
)

type Interface interface {
	Create(ctx context.Context, u model.User) error
	GetByID(ctx context.Context, id int) (model.User, error)
	List(ctx context.Context) ([]model.User, error)
}

var _ Interface = (*user)(nil)

type user struct {
	repo   repository.User
	logger logger.Interface
}

func New(repo repository.User, logger logger.Interface) *user {
	return &user{
		repo:   repo,
		logger: logger,
	}
}

func (s user) Create(ctx context.Context, u model.User) error {
	return s.repo.Create(ctx, u)
}

func (s user) GetByID(ctx context.Context, id int) (model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s user) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}
