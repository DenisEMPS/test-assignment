package suite

import (
	"log/slog"
	"os"
	"testing"

	"github.com/DenisEMPS/test-assignment/internal/config"
)

type Suite struct {
	*testing.T
	Cfg *config.Config
	Log *slog.Logger
}

func New(t *testing.T) *Suite {
	t.Helper()
	t.Parallel()

	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	cfg := config.MustLoadByPath("../../tests/suite/config.yml")

	return &Suite{
		T:   t,
		Cfg: cfg,
		Log: log,
	}
}
