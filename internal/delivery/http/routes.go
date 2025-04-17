package http

import "github.com/gin-gonic/gin"

type Handler struct {
	authService AuthService
}

func NewHandler(authService AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()

	auth := r.Group("/auth")
	{
		auth.POST("/sign-in", h.GetTokens)
		auth.POST("/refresh", h.RefreshTokens)
	}

	return r
}
