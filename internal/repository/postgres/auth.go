package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrTokenDoesNotExists = errors.New("token does not exists")
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuth(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) SaveRefreshTokenRecord(ctx context.Context, tokenDetails *domain.RefreshTokenRecord) error {
	const op = "AuthPostgres.SaveRefreshToken"

	query := `
		INSERT INTO refresh_sessions (user_id, access_uuid, refresh_hash, ip, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query, tokenDetails.UserID, tokenDetails.AccessUUID, tokenDetails.Hash, tokenDetails.UserIP, tokenDetails.ExpiresAt, tokenDetails.CreatedAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *AuthPostgres) GetRefreshTokenRecord(ctx context.Context, userID uuid.UUID, accessID uuid.UUID) (*domain.TokenRefreshDAO, error) {
	const op = "AuthPostgres.GetRefreshToken"

	query := `
		SELECT refresh_hash, expires_at, ip 
		FROM refresh_sessions
		WHERE user_id = $1 AND access_uuid = $2 
	`

	var tokenDetails domain.TokenRefreshDAO

	err := r.db.QueryRowContext(ctx, query, userID, accessID).Scan(&tokenDetails.Hash, &tokenDetails.ExpiresAt, &tokenDetails.UserIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenDoesNotExists)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &tokenDetails, nil
}

func (r *AuthPostgres) DeleteRefreshTokenRecord(ctx context.Context, userID, accessID uuid.UUID) error {
	const op = "AuthPostgres.DeleteRefreshToken"

	query := `
		DELETE FROM refresh_sessions
		WHERE user_id = $1 AND access_uuid = $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, accessID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
