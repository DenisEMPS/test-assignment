package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/DenisEMPS/test-assignment/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthService interface {
	GenerateTokens(ctx context.Context, userID uuid.UUID, userIP string) (*domain.TokenPairResponse, error)
	RefreshTokens(ctx context.Context, tokenPair *domain.RefreshTokensRequest, userIP string) (*domain.TokenPairResponse, error)
}

func (h *Handler) GenerateTokens(c *gin.Context) {
	idParam, ok := c.GetQuery(qUserID)
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, invalidID)
		return
	}

	userID, err := uuid.Parse(idParam)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, invalidID)
		return
	}

	userIP := c.ClientIP()
	if userIP == empty {
		newErrorResponse(c, http.StatusBadRequest, invalidReq)
		return
	}

	tokens, err := h.authService.GenerateTokens(context.Background(), userID, userIP)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, internalError)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) RefreshTokens(c *gin.Context) {
	var tokenPair *domain.RefreshTokensRequest

	if err := c.ShouldBindJSON(&tokenPair); err != nil {
		newErrorResponse(c, http.StatusBadRequest, invalidReq)
		return
	}

	if tokenPair.Access == empty || tokenPair.Refresh == empty {
		newErrorResponse(c, http.StatusBadRequest, invalidReq)
		return
	}

	userIP := c.ClientIP()
	if userIP == empty {
		newErrorResponse(c, http.StatusBadRequest, invalidReq)
		return
	}

	newTokenPair, err := h.authService.RefreshTokens(context.Background(), tokenPair, userIP)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			newErrorResponse(c, http.StatusUnauthorized, unauthorized)
			return
		} else if errors.Is(err, auth.ErrTokenDoesNotExists) {
			newErrorResponse(c, http.StatusUnauthorized, unauthorized)
			return
		} else if errors.Is(err, auth.ErrTokensNotIdentical) {
			newErrorResponse(c, http.StatusUnauthorized, unauthorized)
			return
		} else if errors.Is(err, auth.ErrRefreshTokenExpired) {
			newErrorResponse(c, http.StatusUnauthorized, unauthorized)
			return
		} else {
			newErrorResponse(c, http.StatusInternalServerError, internalError)
			return
		}
	}

	c.JSON(http.StatusOK, newTokenPair)
}
