package http

import "github.com/gin-gonic/gin"

type errMsg struct {
	Error string `json:"error"`
}

func errorMsg(c *gin.Context, code int, err error) {
	c.JSON(code, errMsg{err.Error()})
}
