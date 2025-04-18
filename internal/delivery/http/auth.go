package http

import (
	"context"
	"net/http"

	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthService interface {
	GenerateTokens(ctx context.Context, userID uuid.UUID, userIP string) (*domain.TokenPairResponse, error)
	RefreshTokens(ctx context.Context, tokenPair *domain.RefreshTokenRequest, userIP string) (*domain.TokenPairResponse, error)
}

const (
	qUserID = "user_id"
)

func (h *Handler) GenerateTokens(c *gin.Context) {
	userIDStr, ok := c.GetQuery(qUserID)
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "id param required")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	userIP := c.ClientIP()
	if userIP == "" {
		newErrorResponse(c, http.StatusBadRequest, "invalid request params")
		return
	}

	tokens, err := h.authService.GenerateTokens(context.Background(), userID, userIP)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "internal error")
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) RefreshTokens(c *gin.Context) {
	var tokenPair *domain.RefreshTokenRequest
	if err := c.BindJSON(&tokenPair); err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "invalid request params")
		return
	}

	userIP := c.ClientIP()
	if userIP == "" {
		newErrorResponse(c, http.StatusUnauthorized, "invalid request params")
		return
	}

	newTokenPair, err := h.authService.RefreshTokens(context.Background(), tokenPair, userIP)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "internal error")
		return
	}

	c.JSON(http.StatusOK, newTokenPair)
}
