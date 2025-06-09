package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"Payment-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	stripeWebhook "github.com/stripe/stripe-go/v74/webhook"
)

// WebhookHandler обрабатывает Stripe-webhook
type WebhookHandler struct {
	webhookSecret  string
	pmService      service.PaymentMethodService
	paymentService service.PaymentService
}

// NewWebhookHandler конструктор
func NewWebhookHandler(secret string, pmSvc service.PaymentMethodService, paySvc service.PaymentService) *WebhookHandler {
	return &WebhookHandler{
		webhookSecret:  secret,
		pmService:      pmSvc,
		paymentService: paySvc,
	}
}

// HandleWebhook — POST /stripe/webhook
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "read error: " + err.Error()})
		return
	}

	// 📦 Логируем "сырое" тело запроса
	log.Printf("📦 Raw payload: %s", string(payload))

	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := stripeWebhook.ConstructEventWithOptions(payload, sigHeader, h.webhookSecret, stripeWebhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})

	if err != nil {
		log.Printf("❌ Signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature: " + err.Error()})
		return
	}

	switch event.Type {
	case "setup_intent.succeeded":
		var si stripe.SetupIntent
		if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
			log.Printf("❌ Failed to parse setup_intent: %v", err)
			break
		}

		userID := si.Metadata["user_id"]
		pmID := si.PaymentMethod.ID

		if userID != "" && pmID != "" {
			_, err := h.pmService.RetrieveAndSavePaymentMethod(c.Request.Context(), userID, pmID)
			if err != nil {
				log.Printf("⚠️ Failed to save card for user %s: %v", userID, err)
			} else {
				log.Printf("✅ Card saved: user_id=%s, pm_id=%s", userID, pmID)
			}
		} else {
			log.Println("⚠️ Missing user_id or pmID in setup_intent")
		}

	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("❌ Failed to parse payment_intent.succeeded: %v", err)
		}

	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("❌ Failed to parse payment_intent.canceled: %v", err)
		}

	default:
		log.Printf("ℹ️ Unhandled event type: %s", event.Type)
	}

	c.Status(http.StatusOK)
}
