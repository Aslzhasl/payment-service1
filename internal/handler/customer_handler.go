package handler

import (
	"context"
	"net/http"

	"Payment-service/internal/service"
	"Payment-service/internal/userclient"
	"github.com/gin-gonic/gin"
)

// CustomerHandler держит зависимости
type CustomerHandler struct {
	svc        service.CustomerService
	userClient *userclient.Client
}

// NewCustomerHandler конструктор
func NewCustomerHandler(svc service.CustomerService, uc *userclient.Client) *CustomerHandler {
	return &CustomerHandler{svc: svc, userClient: uc}
}

// CreateCustomer — POST /api/v1/pay/customers (без тела запроса)
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	email := c.GetString("userEmail")

	user, err := h.userClient.GetByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch user: " + err.Error()})
		return
	}

	stripeID, err := h.svc.EnsureCustomer(context.Background(), user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot ensure Stripe customer: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer_id": stripeID})
}
