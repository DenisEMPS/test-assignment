package postgres

import "github.com/jmoiron/sqlx"

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuth(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) Login() {

}

func (r *AuthPostgres) Refresh() {

}
