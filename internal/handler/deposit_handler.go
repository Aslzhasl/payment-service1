// internal/handler/deposit_handler.go
package handler

import (
	"net/http"

	"Payment-service/internal/service"
	"Payment-service/internal/userclient"

	"github.com/gin-gonic/gin"
)

// DepositHandler держит зависимости для операций с депозитами.
type DepositHandler struct {
	svc        service.DepositService
	custSvc    service.CustomerService
	userClient *userclient.Client
}

// NewDepositHandler конструктор
func NewDepositHandler(
	svc service.DepositService,
	custSvc service.CustomerService,
	userClient *userclient.Client,
) *DepositHandler {
	return &DepositHandler{svc: svc, custSvc: custSvc, userClient: userClient}
}

// CreateDepositRequest — payload для POST /deposits
// booking_id, listing_id, amount и currency приходят из клиента.
// userID и customerID получаем из контекста и CustomerService соответственно.
type CreateDepositRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
	ListingID string `json:"listing_id" binding:"required"`
	Amount    int64  `json:"amount" binding:"required,gt=0"`
	Currency  string `json:"currency" binding:"required"`
}

// CreateDepositResponse — ответ на CREATE, содержит client_secret и deposit_id
type CreateDepositResponse struct {
	ClientSecret string `json:"client_secret"`
	DepositID    string `json:"deposit_id"`
}

// CreateDeposit обрабатывает POST /api/v1/pay/deposits
func (h *DepositHandler) CreateDeposit(c *gin.Context) {
	// 1) Получаем email из JWT, middleware написал ранее
	email := c.GetString("userEmail")

	// 2) Достаём userID из User-service
	user, err := h.userClient.GetByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch user: " + err.Error()})
		return
	}

	// 3) Проверяем или создаём Stripe Customer
	stripeCustID, err := h.custSvc.EnsureCustomer(c.Request.Context(), user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot ensure customer: " + err.Error()})
		return
	}

	// 4) Парсим тело запроса
	var req CreateDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5) Авторизуем депозит
	clientSecret, depositID, err := h.svc.AuthorizeDeposit(
		c.Request.Context(),
		stripeCustID,
		user.ID,
		req.BookingID,
		req.ListingID,
		req.Currency,
		req.Amount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6) Возвращаем клиенту данные для подтверждения
	resp := CreateDepositResponse{ClientSecret: clientSecret, DepositID: depositID}
	c.JSON(http.StatusOK, resp)
}

// CaptureRefundRequest — payload для capture/refund операций
type CaptureRefundRequest struct {
	DepositID string `json:"deposit_id" binding:"required"`
}

// CaptureDeposit обрабатывает POST /api/v1/pay/deposits/capture
func (h *DepositHandler) CaptureDeposit(c *gin.Context) {
	var req CaptureRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CaptureDeposit(c.Request.Context(), req.DepositID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// RefundDeposit обрабатывает POST /api/v1/pay/deposits/refund
func (h *DepositHandler) RefundDeposit(c *gin.Context) {
	var req CaptureRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RefundDeposit(c.Request.Context(), req.DepositID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
