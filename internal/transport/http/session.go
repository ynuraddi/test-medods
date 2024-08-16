package http

import (
	"context"
	"database/sql"
	"errors"
	"medods/internal/model"
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

type updateSessionRequest struct {
	ID         int    `json:"id" binding:"required"`
	UserID     int    `json:"user_id"`
	ATokenID   string `json:"access_token_id"`
	RTokenHash string `json:"refresh_token_hash"`
	IP         string `json:"ip"`
	CreatedAT  int64  `json:"created_at"`
	Version    int64  `json:"version"`
}

// this route only for testing and have not validation
func (h sessionRoutes) updateSession(c *gin.Context) {
	var req updateSessionRequest
	if err := c.BindJSON(&req); err != nil {
		errorMsg(c, http.StatusBadRequest, err)
		return
	}

	input := model.Session{
		ID:         req.ID,
		UserID:     req.UserID,
		ATokenID:   req.ATokenID,
		RTokenHash: req.RTokenHash,
		IP:         req.IP,
		CreatedAt:  req.CreatedAT,
		Version:    req.Version,
	}

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	updatedSession, err := h.sessionService.Update(ctx, input)
	if errors.Is(err, sql.ErrNoRows) {
		errorMsg(c, http.StatusNotFound, err)
		return
	} else if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, updatedSession)
}

func (h sessionRoutes) listSession(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	users, err := h.sessionService.List(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		c.Status(http.StatusNoContent)
		return
	} else if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}
