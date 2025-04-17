package auth

import "log/slog"

type AuthRepository interface {
	Login()
	Refresh()
}

type Auth struct {
	log      *slog.Logger
	authRepo AuthRepository
}

func New(authRepo AuthRepository, log *slog.Logger) *Auth {
	return &Auth{
		authRepo: authRepo,
		log:      log,
	}
}

func (s *Auth) Login() {}

func (s *Auth) Refresh() {}
