package http

import "github.com/gin-gonic/gin"

type AuthService interface {
	Login()
	Refresh()
}

func (h *Handler) Login(c *gin.Context) {

}

func (h *Handler) Refresh(c *gin.Context) {

}
