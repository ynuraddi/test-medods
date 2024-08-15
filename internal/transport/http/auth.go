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

func (h authRoutes) login(c *gin.Context) {
	stringID := c.Param("user_id")
	userID, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("failed ot convert user_id[%s]: %s", stringID, err.Error())})
		return
	}

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	user, err := h.userService.GetByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNoContent, gin.H{"message": "user not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	aToken, rToken, err := h.authService.CreateSession(ctx, user.ID, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (h authRoutes) refresh(c *gin.Context) {
	aToken := c.GetHeader("Authorization")
	aToken = strings.TrimPrefix(aToken, "Bearer ")
	if len(aToken) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token is empty"})
		return
	}

	var req refreshRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.RefreshToken) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token is empty"})
		return
	}
	rToken := req.RefreshToken

	ctx, cancel := context.WithTimeout(c.Copy(), 3*time.Second)
	defer cancel()

	aToken, rToken, err := h.authService.RefreshSession(ctx, aToken, rToken)
	if errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrSignatureInvalid) ||
		errors.Is(err, jwt.ErrTokenMalformed) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, refreshResponse{
		AccessToken:  aToken,
		RefreshToken: rToken,
	})
}
