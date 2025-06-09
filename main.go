package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"

	"github.com/gin-gonic/gin"

	"Payment-service/internal/config"
	"Payment-service/internal/routes"
	"Payment-service/internal/storage"
)

func main() {
	// 1) Загружаем конфиг
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2) Инициализируем подключение к БД
	store, err := storage.InitStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db init error: %v", err)
	}
	defer store.Close()

	// 3) Настраиваем HTTP и роуты
	r := gin.Default()
	// Теперь передаём webhookSecret и stripeAPIKey
	routes.RegisterAll(
		r,
		store,
		cfg.StripeWebhookSecret,
		cfg.StripeSecretKey,
		cfg.JWTSecret,
		cfg.UserServiceURL,
	)
	fmt.Println(">>> STRIPE_WEBHOOK_SECRET:", cfg.StripeWebhookSecret)
	// 4) Запуск сервера
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("🚀 Payment-service listening on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
