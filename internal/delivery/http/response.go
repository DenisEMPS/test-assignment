package http

import (
	"log"

	"github.com/gin-gonic/gin"
)

type Response struct {
	message string `json:"message"`
}

func newErrorResponse(c *gin.Context, code int, message string) {
	log.Print(message)
	c.AbortWithStatusJSON(code, Response{
		message: message,
	})
}
