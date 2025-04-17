package handler

import "github.com/gin-gonic/gin"

type AuthService interface {
	Login()
	Refresh()
}

type Handler struct {
	authService AuthService
}

func New(authService AuthService) *Handler {
	return &Handler{authService: authService}
}

func InitRoutes(h *Handler) *gin.Engine {
	r := gin.New()

	auth := r.Group("/auth")
	{
		auth.POST("/sign-in", h.Login)
		auth.POST("/refresh", h.Refresh)
	}

	return r
}
