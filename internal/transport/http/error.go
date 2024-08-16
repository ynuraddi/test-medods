package http

import "github.com/gin-gonic/gin"

func errorMsg(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{"error": err.Error()})
}
