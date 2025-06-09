package handler

import (
	"net/http"

	"Payment-service/internal/service"
	"Payment-service/internal/userclient"

	"github.com/gin-gonic/gin"
	stripepkg "github.com/stripe/stripe-go/v74"
)

// PaymentMethodHandler держит зависимости
// svc        – сервис для управления SetupIntent
// custSvc    – сервис для управления Stripe Customer
// userClient – HTTP-клиент User-service
type PaymentMethodHandler struct {
	svc        service.PaymentMethodService
	custSvc    service.CustomerService
	userClient *userclient.Client
}

// NewPaymentMethodHandler конструктор
func NewPaymentMethodHandler(
	svc service.PaymentMethodService,
	custSvc service.CustomerService,
	userClient *userclient.Client,
) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc, custSvc: custSvc, userClient: userClient}
}

// CreateSetupIntentRequest — payload для /setup-intents
type CreateSetupIntentRequest struct {
	Usage string `json:"usage" binding:"required"` // "off_session" или "on_session"
}

// CreateSetupIntent обрабатывает POST /api/v1/pay/setup-intents
func (h *PaymentMethodHandler) CreateSetupIntent(c *gin.Context) {
	// 1) Email из JWT, middleware положил его в контекст
	email := c.GetString("userEmail")

	// 2) Получить userID из User-service
	user, err := h.userClient.GetByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch user: " + err.Error()})
		return
	}

	// 3) Убедиться, что есть Stripe-Customer
	stripeCustomerID, err := h.custSvc.EnsureCustomer(c.Request.Context(), user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot ensure customer: " + err.Error()})
		return
	}

	// 4) Прочитать JSON-запрос
	var req CreateSetupIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5) Создать SetupIntent
	clientSecret, err := h.svc.CreateSetupIntent(
		c.Request.Context(),
		stripeCustomerID,
		stripepkg.SetupIntentUsage(req.Usage),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6) Ответ
	c.JSON(http.StatusOK, gin.H{"client_secret": clientSecret})
}

// ListPaymentMethods обрабатывает GET /api/v1/pay/payment-methods
func (h *PaymentMethodHandler) ListPaymentMethods(c *gin.Context) {
	// 1) Email из JWT
	email := c.GetString("userEmail")

	// 2) Получить userID
	user, err := h.userClient.GetByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch user: " + err.Error()})
		return
	}

	// 3) Запросить сохранённые карты
	methods, err := h.svc.ListByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, methods)
}
