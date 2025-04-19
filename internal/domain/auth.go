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
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type RefreshTokensRequest struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type TokenRefreshDAO struct {
	Hash      string    `db:"refresh_hash"`
	UserIP    string    `db:"ip"`
	ExpiresAt time.Time `db:"expires_at"`
}

type RefreshTokenRecord struct {
	Hash       string
	UserID     uuid.UUID
	AccessUUID uuid.UUID
	UserIP     string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}
