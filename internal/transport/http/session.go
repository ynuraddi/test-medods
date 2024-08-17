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

// UpdateSession godoc
//
//	@Summary		Update session
//	@Description	Update session in database. Needed for testing.
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Param			update_request	body	updateSessionRequest	true	"update session request, find session by id and version and update"
//	@Success		200
//	@Failure		400	{object}	errMsg	"Invalid request parameters"
//	@Failure		404	{object}	errMsg	"Not found"
//	@Failure		500	{object}	errMsg	"Internal server error"
//	@Router			/session/update [post]
func (h sessionRoutes) updateSession(c *gin.Context) {
	// this route only for testing and have not validation
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

// ListSession godoc
//
//	@Summary		List session
//	@Description	Show rows of session table from database.
//	@Tags			test
//	@Success		200
//	@Success		204
//	@Failure		500	{object}	errMsg	"Internal server error"
//	@Router			/session/list [get]
func (h sessionRoutes) listSession(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	users, err := h.sessionService.List(ctx)
	if errors.Is(err, sql.ErrNoRows) || len(users) == 0 {
		c.Status(http.StatusNoContent)
		return
	} else if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}
