package http

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"medods/internal/service"
	"medods/internal/service/auth"
	"medods/internal/service/user"
	"medods/pkg/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type authRoutes struct {
	authService auth.Interface
	userService user.Interface
	logger      logger.Interface
}

func newAuthRoutes(l logger.Interface, s *service.Manager) *authRoutes {
	return &authRoutes{
		authService: s.Auth,
		userService: s.User,

		logger: l,
	}
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// CreateSession godoc
//
//	@Summary		Create session
//	@Description	Create session and return new pair access and refresh tokens.
//	@Tags			auth
//	@Produce		json
//	@Param			user_id	path	int	true	"user id"
//	@Success		201
//	@Failure		400	{object}	errMsg	"Invalid request parameters"
//	@Failure		404	{object}	errMsg	"User not found"
//	@Failure		500	{object}	errMsg	"Internal server error"
//	@Router			/auth/login/{user_id} [get]
func (h authRoutes) login(c *gin.Context) {
	stringID := c.Param("user_id")
	userID, err := strconv.Atoi(stringID)
	if err != nil {
		errorMsg(c, http.StatusBadRequest, fmt.Errorf("failed ot convert user_id[%s]: %s", stringID, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	user, err := h.userService.GetByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	aToken, rToken, err := h.authService.CreateSession(ctx, user.ID, c.ClientIP())
	if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		AccessToken:  aToken,
		RefreshToken: rToken,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshSession godoc
//
//	@Summary		Refresh session
//	@Description	Refresh session and return new pair access and refresh tokens.
//	@Security		BearerAuth
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh_token	body	refreshRequest	true	"refresh token for refresh session"
//	@Success		200
//	@Failure		400	{object}	errMsg	"Invalid request parameters"
//	@Failure		401	{object}	errMsg	"Unauthorized - invalid tokens"
//	@Failure		500	{object}	errMsg	"Internal server error"
//	@Router			/auth/refresh [post]
func (h authRoutes) refresh(c *gin.Context) {
	// не сдела middleware потому что подумал что тут логично пропускать даже expire aToken
	aToken := c.GetHeader("Authorization")
	aToken = strings.TrimPrefix(aToken, "Bearer ")
	if len(aToken) == 0 {
		errorMsg(c, http.StatusUnauthorized, fmt.Errorf("authorization token is empty"))
		return
	}

	var req refreshRequest
	if err := c.BindJSON(&req); err != nil {
		errorMsg(c, http.StatusBadRequest, err)
		return
	}
	rToken := req.RefreshToken

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	aToken, rToken, err := h.authService.RefreshSession(ctx, aToken, rToken, c.ClientIP())
	if errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrSignatureInvalid) ||
		errors.Is(err, jwt.ErrTokenMalformed) ||
		errors.Is(err, auth.ErrValidationFailed) {
		errorMsg(c, http.StatusUnauthorized, err)
		return
	} else if err != nil {
		errorMsg(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, refreshResponse{
		AccessToken:  aToken,
		RefreshToken: rToken,
	})
}
