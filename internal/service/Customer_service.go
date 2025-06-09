// internal/service/customer_service.go
package service

import (
	"Payment-service/internal/userclient"
	"context"

	"Payment-service/internal/repository"
	"Payment-service/internal/stripeadapter"
)

// CustomerService defines business logic around Stripe Customers.
type CustomerService interface {
	// EnsureCustomer checks if a Stripe Customer exists for userID; if not, creates it.
	EnsureCustomer(ctx context.Context, userID, email string) (string, error)
}

// customerService is a concrete implementation of CustomerService.
type customerService struct {
	repo       repository.CustomerRepo
	stripe     *stripeadapter.Client
	userClient *userclient.Client
}

// NewCustomerService constructs a CustomerService.
func NewCustomerService(repo repository.CustomerRepo, client *stripeadapter.Client, userclient *userclient.Client) CustomerService {
	return &customerService{repo: repo, stripe: client, userClient: userclient}
}

// EnsureCustomer checks for existing Customer in DB, creates in Stripe if missing.
func (s *customerService) EnsureCustomer(ctx context.Context, userID, email string) (string, error) {
	// 1) Try to fetch from DB
	id, err := s.repo.GetCustomerByUserID(ctx, userID)
	if err == nil && id != "" {
		return id, nil
	}

	// 2) Not found: create in Stripe
	newID, err := s.stripe.CreateCustomer(ctx, email, userID)
	if err != nil {
		return "", err
	}

	// 3) Persist mapping in DB
	if err := s.repo.CreateCustomer(ctx, userID, email, newID); err != nil {
		return "", err
	}
	return newID, nil
}

func (s *customerService) UserClient() *userclient.Client {
	return s.userClient
}
