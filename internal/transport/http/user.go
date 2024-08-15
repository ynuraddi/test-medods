package http

import (
	"context"
	"database/sql"
	"errors"
	"medods/internal/model"
	"medods/internal/service"
	"medods/internal/service/user"
	"medods/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type userRoutes struct {
	userService user.Interface
	logger      logger.Interface
}

func newUserRoutes(l logger.Interface, s *service.Manager) *userRoutes {
	return &userRoutes{
		userService: s.User,
		logger:      l,
	}
}

type createUserRequest struct {
	Email string `json:"email" binding:"required,email" example:"mock@gmail.com"`
}

func (h userRoutes) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	if err := h.userService.Create(ctx, model.User{
		Email: req.Email,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h userRoutes) listUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	users, err := h.userService.List(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		c.Status(http.StatusNoContent)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
