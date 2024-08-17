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

// CreateUser godoc
//
//	@Summary		Create user
//	@Description	Create user, needed for testing with few users
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Param			create_request	body	createUserRequest	true	"create user request, email of user" Format(email)
//	@Success		201
//	@Failure		400	{object}	errMsg	"Invalid request parameters"
//	@Failure		500	{object}	errMsg	"Internal server error"
//	@Router			/user/create [post]
func (h userRoutes) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.BindJSON(&req); err != nil {
		errorMsg(c, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	if err := h.userService.Create(ctx, model.User{
		Email: req.Email,
	}); err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// ListUsers godoc
//
//	@Summary		List user
//	@Description	Show rows of users table from database.
//	@Tags			test
//	@Success		200
//	@Success		204
//	@Failure		500	{object}	errMsg	"Internal server error"
//	@Router			/user/list [get]
func (h userRoutes) listUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	users, err := h.userService.List(ctx)
	if errors.Is(err, sql.ErrNoRows) || len(users) == 0 {
		c.Status(http.StatusNoContent)
		return
	} else if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}
