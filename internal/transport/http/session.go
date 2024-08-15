package http

import (
	"context"
	"database/sql"
	"errors"
	"medods/internal/service"
	"medods/internal/service/session"
	"medods/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type sessionRoutes struct {
	sessionService session.Interface
	logger         logger.Interface
}

func newSessionRoutes(l logger.Interface, s *service.Manager) *sessionRoutes {
	return &sessionRoutes{
		sessionService: s.Session,
		logger:         l,
	}
}

func (h sessionRoutes) listSession(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	users, err := h.sessionService.List(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		c.Status(http.StatusNoContent)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
