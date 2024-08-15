package service

import (
	"medods/config"
	"medods/internal/repository"
	"medods/internal/service/auth"
	"medods/internal/service/jwt"
	"medods/internal/service/session"
	"medods/internal/service/user"
	"medods/pkg/logger"
)

type Manager struct {
	Auth    auth.Interface
	User    user.Interface
	Session session.Interface
}

func New(cfg *config.Config, repo *repository.Manager, l logger.Interface) *Manager {
	userService := user.New(repo.User, l)
	sessionService := session.New(repo.Session, l)
	jwtMaker := jwt.New([]byte(cfg.JWT.SecretKey), l)
	authService := auth.New(sessionService, jwtMaker, l, false)

	return &Manager{
		Auth: authService,
		User: userService,
	}
}
