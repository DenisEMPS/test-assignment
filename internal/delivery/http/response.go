package http

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, Response{
		Message: message,
	})
}
