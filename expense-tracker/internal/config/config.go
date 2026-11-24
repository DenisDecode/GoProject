// Package config читает параметры запуска из переменных окружения
// (со значениями по умолчанию для локальной разработки).
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	DBPath            string
	TemplatesDir      string
	StaticDir         string
	MigrationsDir     string
	SessionTTL        time.Duration
	CookieSecure      bool
	EnableDebugRoutes bool
}

func Load() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		DBPath:            getEnv("DB_PATH", "expenses.db"),
		TemplatesDir:      getEnv("TEMPLATES_DIR", "web/templates"),
		StaticDir:         getEnv("STATIC_DIR", "web/static"),
		MigrationsDir:     getEnv("MIGRATIONS_DIR", "migrations"),
		SessionTTL:        7 * 24 * time.Hour,
		CookieSecure:      getBool("COOKIE_SECURE", false),
		EnableDebugRoutes: getBool("ENABLE_DEBUG_ROUTES", true),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
