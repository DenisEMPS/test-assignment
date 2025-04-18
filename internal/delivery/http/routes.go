package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log         *slog.Logger
	authService AuthService
}

func NewHandler(authService AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()

	auth := r.Group("/auth")
	{
		auth.POST("/generate", h.GenerateTokens)
		auth.POST("/refresh", h.RefreshTokens)
	}

	return r
}
