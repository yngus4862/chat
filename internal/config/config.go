package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort string
	WsPort  string

	PGHost string
	PGPort string
	PGUser string
	PGPass string
	PGDB   string
}

func Load() Config {
	return Config{
		AppPort: env("APP_PORT", "8080"),
		WsPort:  env("APP_WS_PORT", "8081"),

		PGHost: env("POSTGRES_HOST", "postgres"),
		PGPort: env("POSTGRES_PORT", "5432"),
		PGUser: env("POSTGRES_USER", "appuser"),
		PGPass: env("POSTGRES_PASSWORD", "appsecret"),
		PGDB:   env("POSTGRES_DB", "chatapp"),
	}
}

func (c Config) PostgresDSN() string {
	// pq driver(dsn) format
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PGHost, c.PGPort, c.PGUser, c.PGPass, c.PGDB,
	)
}

func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}