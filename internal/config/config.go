package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppHTTPAddr   string
	AppWSAddr     string
	AdminHTTPAddr string
	AdminToken    string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	RedisHost string
	RedisPort string
}

func Load() Config {
	cfg := Config{
		AppHTTPAddr:   env("APP_HTTP_ADDR", "0.0.0.0:8080"),
		AppWSAddr:     env("APP_WS_ADDR", "0.0.0.0:8081"),
		AdminHTTPAddr: env("ADMIN_HTTP_ADDR", "127.0.0.1:9099"),
		AdminToken:    env("ADMIN_TOKEN", ""),

		PostgresHost:     env("POSTGRES_HOST", "postgres"),
		PostgresPort:     env("POSTGRES_PORT", "5432"),
		PostgresUser:     env("POSTGRES_USER", "appuser"),
		PostgresPassword: env("POSTGRES_PASSWORD", "appsecret"),
		PostgresDB:       env("POSTGRES_DB", "chatapp"),

		RedisHost: env("REDIS_HOST", "redis"),
		RedisPort: env("REDIS_PORT", "6379"),
	}
	return cfg
}

func (c Config) PostgresURL() string {
	// postgres://user:pass@host:port/db?sslmode=disable
	user := urlEscape(c.PostgresUser)
	pass := urlEscape(c.PostgresPassword)
	host := strings.TrimSpace(c.PostgresHost)
	port := strings.TrimSpace(c.PostgresPort)
	db := strings.TrimSpace(c.PostgresDB)
	if host == "" {
		host = "postgres"
	}
	if port == "" {
		port = "5432"
	}
	if db == "" {
		db = "chatapp"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func (c Config) RedisAddr() string {
	host := strings.TrimSpace(c.RedisHost)
	port := strings.TrimSpace(c.RedisPort)
	if host == "" {
		host = "redis"
	}
	if port == "" {
		port = "6379"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func urlEscape(s string) string {
	// minimal escape to keep DSN safe for typical passwords (optional)
	// pg DSN can accept raw, but we keep this conservative.
	return strings.ReplaceAll(s, "@", "%40")
}
