package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/DenisEMPS/test-assignment/internal/config"
	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/DenisEMPS/test-assignment/internal/repository/postgres"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTokenExpired       = errors.New("token is expired")
	ErrTokenDoesNotExists = errors.New("token does not exists")
	ErrTokenNotIdentical  = errors.New("tokens are not identical")
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, tokenDetails *domain.TokenRefreshDetails) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID, accessID uuid.UUID) (*domain.TokenRefreshDAO, error)
	DeleteRefreshToken(ctx context.Context, userID, accessID uuid.UUID) error
}

type Auth struct {
	cfg      *config.Token
	log      *slog.Logger
	authRepo AuthRepository
}

func New(authRepo AuthRepository, log *slog.Logger, cfg *config.Token) *Auth {
	return &Auth{
		authRepo: authRepo,
		log:      log,
		cfg:      cfg,
	}
}

func (s *Auth) GenerateTokens(ctx context.Context, userID uuid.UUID, userIP string) (*domain.TokenPairResponse, error) {
	const op = "Auth.GenerateTokens"

	log := s.log.With(
		slog.String("op", op),
		slog.Any("user_id", userID),
	)

	accessID := uuid.New()
	accessToken, err := s.GenerateAccessToken(ctx, userID, accessID, userIP, s.cfg.AccessTokenTTL)
	if err != nil {
		log.Error("failed to generate access token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, refreshHash, err := s.GenerateRefreshToken(ctx)
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refTokenDetails := &domain.TokenRefreshDetails{
		TokenHash:  refreshHash,
		UserID:     userID,
		AccessUUID: accessID,
		UserIP:     userIP,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	err = s.authRepo.SaveRefreshToken(ctx, refTokenDetails)
	if err != nil {
		log.Error("failed to save refresh token in database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Auth) GenerateAccessToken(ctx context.Context, userID uuid.UUID, accessID uuid.UUID, userIP string, duration time.Duration) (string, error) {
	const op = "Auth.GenerateAccessToken"

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, domain.UserClaims{
		UserID:     userID,
		UserIP:     userIP,
		AccessUUID: accessID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	})

	tokenString, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}

func (s *Auth) GenerateRefreshToken(ctx context.Context) (string, string, error) {
	const op = "Auth.GenerateRefreshToken"

	bts := make([]byte, 32)
	if _, err := rand.Read(bts); err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	hash, err := bcrypt.GenerateFromPassword(bts, bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return base64.StdEncoding.EncodeToString([]byte(bts)), string(hash), nil
}

func (s *Auth) ParseAccessToken(ctx context.Context, token string) (*domain.UserClaims, error) {
	const op = "Auth.ParseAccessToken"

	var errType error

	tokenParsed, err := jwt.ParseWithClaims(token, &domain.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				err = ErrTokenExpired
			}
		}
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := tokenParsed.Claims.(*domain.UserClaims)
	if !ok {
		return nil, fmt.Errorf("%s, claims not in type", op)
	}

	return claims, errType
}

func (s *Auth) RefreshTokens(ctx context.Context, tokenPair *domain.RefreshTokenRequest, userIP string) (*domain.TokenPairResponse, error) {
	const op = "Auth.RefreshTokens"

	log := s.log.With(
		slog.String("op", op),
	)

	claims, err := s.ParseAccessToken(ctx, tokenPair.AccessToken)
	if err != nil {
		if !errors.Is(err, ErrTokenExpired) {
			log.Warn("invalid token", slog.String("error", err.Error()), slog.Any("user_id", claims.UserID))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	oldRefTokenDetails, err := s.authRepo.GetRefreshToken(ctx, claims.UserID, claims.AccessUUID)
	if err != nil {
		if errors.Is(err, postgres.ErrTokenDoesNotExists) {
			log.Warn("invalid token", slog.String("error", err.Error()), slog.Any("user_id", claims.UserID))
			return nil, fmt.Errorf("%s: %w", op, ErrTokenDoesNotExists)
		}
		log.Error("failed to get refresh token", slog.String("error", err.Error()), slog.Any("user_id", claims.UserID))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	decodedOldRefToken, err := base64.StdEncoding.DecodeString(tokenPair.RefreshToken)
	if err != nil {
		log.Error("failed to decode", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: failed to decode refresh token %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(oldRefTokenDetails.TokenHash), decodedOldRefToken); err != nil {
		log.Warn("failed to compare refresh tokens", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, ErrTokenNotIdentical)
	}

	if oldRefTokenDetails.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("%s: %w", op, ErrTokenExpired)
	}

	if userIP != claims.UserIP {
		fmt.Println("A warning about logging in from a different IP address")
	}

	accessID := uuid.New()
	newAccessToken, err := s.GenerateAccessToken(ctx, claims.UserID, accessID, userIP, s.cfg.AccessTokenTTL)
	if err != nil {
		log.Error("failed to generate access token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newRefreshToken, refreshHash, err := s.GenerateRefreshToken(ctx)
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newRefTokenDetails := &domain.TokenRefreshDetails{
		TokenHash:  refreshHash,
		UserID:     claims.UserID,
		AccessUUID: accessID,
		UserIP:     userIP,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	err = s.authRepo.DeleteRefreshToken(ctx, claims.UserID, claims.AccessUUID)
	if err != nil {
		log.Error("failed to delete old refresh token in database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.authRepo.SaveRefreshToken(ctx, newRefTokenDetails)
	if err != nil {
		log.Error("failed to save refresh token in database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPairResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
