package domain

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type UserClaims struct {
	UserID     uuid.UUID `json:"user_id"`
	AccessUUID uuid.UUID `json:"access_uuid"`
	UserIP     string    `json:"client_ip"`
	jwt.StandardClaims
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenAccessParsed struct {
	UserID     uuid.UUID
	AccessUUID uuid.UUID
	UserIP     string
}

type TokenRefreshDAO struct {
	TokenHash string    `db:"refresh_hash"`
	UserIP    string    `db:"ip"`
	ExpiresAt time.Time `db:"expires_at"`
}

type TokenRefreshDetails struct {
	TokenHash  string
	UserID     uuid.UUID
	AccessUUID uuid.UUID
	UserIP     string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}
