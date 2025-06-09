package repository

import "context"

type CustomerRepo interface {
	CreateCustomer(ctx context.Context, userID, email, stripeID string) error
	GetCustomerByUserID(ctx context.Context, userID string) (string, error)
}
