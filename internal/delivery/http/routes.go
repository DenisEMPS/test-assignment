package http

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService AuthService
}

func NewHandler(authService AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/generate", h.GenerateTokens)
		auth.POST("/refresh", h.RefreshTokens)
	}

	return r
}
