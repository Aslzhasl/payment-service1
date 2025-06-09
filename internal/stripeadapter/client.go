// internal/stripe/client.go
package stripeadapter

import (
	"context"
	"github.com/stripe/stripe-go/v74/paymentmethod"

	stripepkg "github.com/stripe/stripe-go/v74"
	stripeCustomer "github.com/stripe/stripe-go/v74/customer"
	stripePayment "github.com/stripe/stripe-go/v74/paymentintent"
	stripeSetup "github.com/stripe/stripe-go/v74/setupintent"
)

// Client wraps Stripe operations needed by the Payment-service.
type Client struct{}

// NewClient sets the Stripe API key and returns a new Client.
func NewClient(apiKey string) *Client {
	stripepkg.Key = apiKey
	return &Client{}
}

// CreateCustomer creates a Stripe Customer with given email and metadata.user_id.
// Returns the Stripe Customer ID (cus_...).
func (c *Client) CreateCustomer(ctx context.Context, email, userID string) (string, error) {
	params := &stripepkg.CustomerParams{
		Email: stripepkg.String(email),
	}
	params.AddMetadata("user_id", userID)
	cust, err := stripeCustomer.New(params)
	if err != nil {
		return "", err
	}
	return cust.ID, nil
}

// CreateSetupIntent issues a SetupIntent to save and verify a card for a Customer.
// Usage should be one of stripe.SetupIntentUsageOffSession or stripe.SetupIntentUsageOnSession.
func (c *Client) CreateSetupIntent(ctx context.Context, customerID string, usage stripepkg.SetupIntentUsage) (*stripepkg.SetupIntent, error) {
	params := &stripepkg.SetupIntentParams{
		Customer:           stripepkg.String(customerID),
		PaymentMethodTypes: stripepkg.StringSlice([]string{"card"}),
		Usage:              stripepkg.String(string(usage)),
	}
	si, err := stripeSetup.New(params)
	if err != nil {
		return nil, err
	}
	return si, nil
}

// CreatePaymentIntent creates a manual-capture PaymentIntent for a given Customer.
// amount is in the smallest currency unit (cents or kopecks).
// bookingID and userID are added to metadata for reference.
func (c *Client) CreatePaymentIntent(ctx context.Context, customerID string, amount int64, currency, bookingID, userID string, listingID string) (*stripepkg.PaymentIntent, error) {
	params := &stripepkg.PaymentIntentParams{
		Amount:             stripepkg.Int64(amount),
		Currency:           stripepkg.String(currency),
		Customer:           stripepkg.String(customerID),
		CaptureMethod:      stripepkg.String("manual"),
		PaymentMethodTypes: stripepkg.StringSlice([]string{"card"}),
	}
	params.AddMetadata("booking_id", bookingID)
	params.AddMetadata("user_id", userID)
	params.AddMetadata("listing_id", listingID)

	pi, err := stripePayment.New(params)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// CapturePaymentIntent captures (finalizes) a previously created & confirmed PaymentIntent.
// Returns the updated PaymentIntent with status "succeeded" on success.
func (c *Client) CapturePaymentIntent(ctx context.Context, paymentIntentID string) (*stripepkg.PaymentIntent, error) {
	pi, err := stripePayment.Capture(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// CancelPaymentIntent cancels a previously authorized PaymentIntent, releasing the hold.
// Returns the updated PaymentIntent with status "canceled" on success.
func (c *Client) CancelPaymentIntent(ctx context.Context, paymentIntentID string) (*stripepkg.PaymentIntent, error) {
	pi, err := stripePayment.Cancel(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return pi, nil
}
func (c *Client) RetrieveCard(pmID string) (*stripepkg.PaymentMethod, error) {
	pm, err := paymentmethod.Get(pmID, nil)
	if err != nil {
		return nil, err
	}
	return pm, nil
}
