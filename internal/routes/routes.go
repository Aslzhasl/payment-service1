package routes

import (
	"Payment-service/internal/handler"
	"Payment-service/internal/middleware"
	"Payment-service/internal/service"
	"Payment-service/internal/storage"
	"Payment-service/internal/stripeadapter"
	"Payment-service/internal/userclient"

	"github.com/gin-gonic/gin"
)

// RegisterAll инициализирует все маршруты и зависимости.
// Теперь принимает jwtSecret и userServiceURL для middleware и userclient.
func RegisterAll(
	r *gin.Engine,
	db *storage.Store,
	webhookSecret,
	stripeAPIKey,
	jwtSecret,
	userServiceURL string,
) {
	// 1) Stripe client
	stripeClient := stripeadapter.NewClient(stripeAPIKey)

	// 2) Репозитории
	custRepo := db // Store реализует repository.CustomerRepo
	pmRepo := db   // Store реализует repository.PaymentMethodRepo
	piRepo := db   // Store реализует repository.PaymentIntentRepo

	// 3) User-client
	userClient := userclient.New(userServiceURL)

	// 4) Сервисы
	custSvc := service.NewCustomerService(custRepo, stripeClient, userClient)
	pmSvc := service.NewPaymentMethodService(pmRepo, stripeClient)
	paySvc := service.NewPaymentService(piRepo, stripeClient)

	// 5) Хендлеры
	custH := handler.NewCustomerHandler(custSvc, userClient)
	pmH := handler.NewPaymentMethodHandler(pmSvc, custSvc, userClient)
	payH := handler.NewPaymentHandler(paySvc)
	whH := handler.NewWebhookHandler(webhookSecret, pmSvc, paySvc)

	// 6) Группа с JWT-мидлвэром
	api := r.Group("/api/v1/pay")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("/customers", custH.CreateCustomer)
		api.POST("/setup-intents", pmH.CreateSetupIntent)
		api.GET("/payment-methods", pmH.ListPaymentMethods)
		api.POST("/payment-intents", payH.CreatePaymentIntent)
		api.POST("/payment-intents/capture", payH.CapturePayment)
		api.POST("/payment-intents/cancel", payH.CancelPayment)
	}

	// Webhook
	r.POST("/stripe/webhook", whH.HandleWebhook)
}
