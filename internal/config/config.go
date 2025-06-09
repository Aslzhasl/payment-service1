package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

// Config хранит все нужные настройки из окружения
type Config struct {
	Port                int    ` env:"PORT,required"`
	DatabaseURL         string `env:"DATABASE_URL,required"`
	JWTSecret           string `env:"JWT_SECRET,required"`
	UserServiceURL      string `env:"USER_SERVICE_URL,required" ` // ← вот это поле
	StripeSecretKey     string `env:"STRIPE_SECRET_KEY,required"`
	StripeWebhookSecret string `env:"STRIPE_WEBHOOK_SECRET,required"`
}

// Load читает .env и парсит переменнfunc
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL must be set")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	cfg.Port = port

	cfg.StripeSecretKey = os.Getenv("STRIPE_SECRET_KEY")
	if cfg.StripeSecretKey == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY must be set")
	}

	cfg.StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")

	// ✅ Вот это добавь
	cfg.UserServiceURL = os.Getenv("USER_SERVICE_URL")
	if cfg.UserServiceURL == "" {
		return nil, fmt.Errorf("USER_SERVICE_URL must be set")
	}

	return cfg, nil
}
