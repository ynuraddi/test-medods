package http

import (
	"medods/internal/service"
	"medods/pkg/logger"

	_ "medods/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func NewRouter(servise *service.Manager, l logger.Interface) *gin.Engine {
	authRoutes := newAuthRoutes(l, servise)
	userRoutes := newUserRoutes(l, servise)
	sessionRoutes := newSessionRoutes(l, servise)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.GET("/login/:user_id", authRoutes.login)
	auth.POST("/refresh", authRoutes.refresh)

	user := api.Group("/user")
	user.POST("/create", userRoutes.createUser)
	user.GET("/list", userRoutes.listUser)

	session := api.Group("/session")
	session.GET("/list", sessionRoutes.listSession)
	// вообще по хорошему /:id/update но ручка просто для теста
	session.POST("/update", sessionRoutes.updateSession)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
