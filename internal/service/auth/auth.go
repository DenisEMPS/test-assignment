package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/DenisEMPS/test-assignment/internal/config"
	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/DenisEMPS/test-assignment/internal/repository/postgres"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTokenExpired        = errors.New("token is expired")
	ErrTokenDoesNotExists  = errors.New("token does not exists")
	ErrTokenNotIdentical   = errors.New("tokens are not identical")
	ErrInvalidToken        = errors.New("invalid token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, tokenDetails *domain.RefreshTokenDetails) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID, accessID uuid.UUID) (*domain.TokenRefreshDAO, error)
	DeleteRefreshToken(ctx context.Context, userID, accessID uuid.UUID) error
}

type Auth struct {
	cfg      *config.JWT
	log      *slog.Logger
	authRepo AuthRepository
}

func New(authRepo AuthRepository, log *slog.Logger, cfg *config.JWT) *Auth {
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

	accessToken, err := s.GenerateAccessToken(userID, accessID, userIP, s.cfg.AccessTokenTTL)
	if err != nil {
		log.Error("failed to generate access token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, refreshHash, err := s.GenerateRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokenDetails := &domain.RefreshTokenDetails{
		Hash:       refreshHash,
		UserID:     userID,
		AccessUUID: accessID,
		UserIP:     userIP,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	err = s.authRepo.SaveRefreshToken(ctx, tokenDetails)
	if err != nil {
		log.Error("failed to save refresh token in postgres", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPairResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *Auth) GenerateAccessToken(userID uuid.UUID, accessID uuid.UUID, userIP string, duration time.Duration) (string, error) {
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

func (s *Auth) GenerateRefreshToken() (string, string, error) {
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

func (s *Auth) ParseAccessToken(token string) (*domain.UserClaims, error) {
	const op = "Auth.ParseAccessToken"

	var errType error

	tokenParsed, err := jwt.ParseWithClaims(token, &domain.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors&jwt.ValidationErrorExpired != 0 {
			errType = ErrTokenExpired
		} else {
			return nil, fmt.Errorf("%s invalid token: %w", op, err)
		}
	}

	claims, ok := tokenParsed.Claims.(*domain.UserClaims)
	if !ok {
		return nil, fmt.Errorf("%s: claims not in type", op)
	}

	return claims, errType
}

func (s *Auth) RefreshTokens(ctx context.Context, tokenPair *domain.RefreshTokensRequest, userIP string) (*domain.TokenPairResponse, error) {
	const op = "Auth.RefreshTokens"

	log := s.log.With(
		slog.String("op", op),
	)

	claims, err := s.ParseAccessToken(tokenPair.Access)
	if err != nil {
		if !errors.Is(err, ErrTokenExpired) {
			log.Warn("failed to parse token", slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
		}
	}

	log = log.With(slog.Any("user_id", claims.UserID))

	oldRefreshData, err := s.authRepo.GetRefreshToken(ctx, claims.UserID, claims.AccessUUID)
	if err != nil {
		if errors.Is(err, postgres.ErrTokenDoesNotExists) {
			log.Warn("token does not exists")
			return nil, fmt.Errorf("%s: %w", op, ErrTokenDoesNotExists)
		}
		log.Error("failed to get refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	currRefreshToken, err := base64.StdEncoding.DecodeString(tokenPair.Refresh)
	if err != nil {
		log.Warn("failed to decode input refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: failed to decode input refresh token %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(oldRefreshData.Hash), currRefreshToken); err != nil {
		log.Warn("failed to compare refresh tokens", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, ErrTokenNotIdentical)
	}

	if oldRefreshData.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("%s: %w", op, ErrRefreshTokenExpired)
	}

	if userIP != oldRefreshData.UserIP {
		fmt.Println("WARNING about logging in from a different IP address")
	}

	accessID := uuid.New()
	newAccessToken, err := s.GenerateAccessToken(claims.UserID, accessID, userIP, s.cfg.AccessTokenTTL)
	if err != nil {
		log.Error("failed to generate access token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newRefreshToken, refreshHash, err := s.GenerateRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	newRefreshData := &domain.RefreshTokenDetails{
		Hash:       refreshHash,
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

	err = s.authRepo.SaveRefreshToken(ctx, newRefreshData)
	if err != nil {
		log.Error("failed to save refresh token in database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokenPairResponse{
		Access:  newAccessToken,
		Refresh: newRefreshToken,
	}, nil
}
