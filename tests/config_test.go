package tests

import (
	"os"
	"testing"

	"github.com/yngus4862/chat/internal/config"
)

func TestPostgresURL(t *testing.T) {
	os.Setenv("POSTGRES_HOST", "postgres")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "u")
	os.Setenv("POSTGRES_PASSWORD", "p")
	os.Setenv("POSTGRES_DB", "d")
	cfg := config.Load()
	got := cfg.PostgresURL()
	want := "postgres://u:p@postgres:5432/d?sslmode=disable"
	if got != want {
		t.Fatalf("got=%s want=%s", got, want)
	}
}
