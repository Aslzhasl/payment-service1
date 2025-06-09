package handler

import (
	"context"
	"net/http"

	"Payment-service/internal/service"
	"github.com/gin-gonic/gin"
)

// PaymentHandler держит зависимости
type PaymentHandler struct {
	svc service.PaymentService
}

// NewPaymentHandler конструктор
func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// CreatePaymentRequest — payload для /payment-intents
type CreatePaymentRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	CustomerID    string `json:"customer_id" binding:"required"`
	BookingID     string `json:"booking_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	PaymentMehtod string `json:"payment_mehtod,omitempty"`
}

// CreatePaymentResponse — ответ
type CreatePaymentResponse struct {
	ClientSecret    string `json:"client_secret"`
	PaymentIntentID string `json:"payment_intent_id"`
}

// CreatePaymentIntent — POST /api/v1/pay/payment-intents
func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	secret, piID, err := h.svc.Authorize(context.Background(),
		req.CustomerID, req.UserID, req.BookingID, req.Currency, req.Amount, req.PaymentMehtod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, CreatePaymentResponse{
		ClientSecret:    secret,
		PaymentIntentID: piID,
	})
}

// CapturePaymentRequest — payload для /payment-intents/capture
type CapturePaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
}

// CapturePayment — POST /api/v1/pay/payment-intents/capture
func (h *PaymentHandler) CapturePayment(c *gin.Context) {
	var req CapturePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Capture(context.Background(), req.PaymentIntentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// CancelPayment — POST /api/v1/pay/payment-intents/cancel
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	var req CapturePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Cancel(context.Background(), req.PaymentIntentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
