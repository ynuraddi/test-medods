package service

import (
	"medods/config"
	"medods/internal/repository"
	"medods/internal/service/auth"
	"medods/internal/service/jwt"
	"medods/internal/service/session"
	"medods/internal/service/user"
	"medods/pkg/logger"
	"medods/pkg/smtp"
)

type Manager struct {
	Auth    auth.Interface
	User    user.Interface
	Session session.Interface
	JWT     jwt.Interface
}

func New(cfg *config.Config, repo *repository.Manager, smtp smtp.Interface, l logger.Interface) *Manager {
	userService := user.New(repo.User, l)
	sessionService := session.New(repo.Session, l)
	jwtMaker := jwt.New([]byte(cfg.JWT.SecretKey), l)
	authService := auth.New(sessionService, userService, jwtMaker, smtp, l, false)

	return &Manager{
		Auth:    authService,
		User:    userService,
		Session: sessionService,
		JWT:     jwtMaker,
	}
}
