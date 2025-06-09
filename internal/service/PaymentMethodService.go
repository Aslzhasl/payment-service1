package service

import (
	"context"
	"log"
	"time"

	"Payment-service/internal/repository"
	stripeadapter "Payment-service/internal/stripeadapter"
	stripePkg "github.com/stripe/stripe-go/v74"
)

// PaymentMethodService defines logic for saving and listing cards.
type PaymentMethodService interface {
	// CreateSetupIntent issues a SetupIntent for the given customer and usage.
	CreateSetupIntent(ctx context.Context, customerID string, usage stripePkg.SetupIntentUsage) (string, error)
	// ListByUser retrieves saved cards for a user.
	ListByUser(ctx context.Context, userID string) ([]repository.PaymentMethod, error)
	RetrieveAndSavePaymentMethod(ctx context.Context, userID, pmID string) (repository.PaymentMethod, error)
}

// paymentMethodService is a concrete implementation of PaymentMethodService.
type paymentMethodService struct {
	repo   repository.PaymentMethodRepo
	stripe *stripeadapter.Client
}

// NewPaymentMethodService constructs a PaymentMethodService.
func NewPaymentMethodService(repo repository.PaymentMethodRepo, client *stripeadapter.Client) PaymentMethodService {
	return &paymentMethodService{repo: repo, stripe: client}
}

// CreateSetupIntent returns a client secret to initialize SetupIntent on the frontend.
func (s *paymentMethodService) CreateSetupIntent(ctx context.Context, customerID string, usage stripePkg.SetupIntentUsage) (string, error) {
	si, err := s.stripe.CreateSetupIntent(ctx, customerID, usage)
	if err != nil {
		return "", err
	}
	return si.ClientSecret, nil
}

// ListByUser returns all saved payment methods for the given user.
func (s *paymentMethodService) ListByUser(ctx context.Context, userID string) ([]repository.PaymentMethod, error) {
	return s.repo.ListPaymentMethods(ctx, userID)
}

// internal/service/PaymentMethodService.go
func (s *paymentMethodService) RetrieveAndSavePaymentMethod(ctx context.Context, userID, pmID string) (repository.PaymentMethod, error) {
	log.Printf("🔎 Retrieving card from Stripe: pmID=%s", pmID)

	card, err := s.stripe.RetrieveCard(pmID)
	if err != nil {
		log.Printf("❌ Failed to retrieve card from Stripe: %v", err)
		return repository.PaymentMethod{}, err
	}

	log.Printf("✅ Retrieved card: brand=%s, last4=%s, exp=%02d/%d",
		card.Card.Brand, card.Card.Last4, card.Card.ExpMonth, card.Card.ExpYear,
	)

	pm := repository.PaymentMethod{
		UserID:     userID,
		StripePMID: card.ID,
		Brand:      string(card.Card.Brand),
		Last4:      card.Card.Last4,
		ExpMonth:   int(card.Card.ExpMonth),
		ExpYear:    int(card.Card.ExpYear),
		CreatedAt:  time.Now(),
	}

	log.Printf("💾 Saving card to DB for userID=%s", userID)

	if err := s.repo.SavePaymentMethod(ctx, pm); err != nil {
		log.Printf("❌ Failed to save card to DB: %v", err)
		return repository.PaymentMethod{}, err
	}

	log.Printf("✅ Card saved successfully for userID=%s", userID)

	return pm, nil
}
