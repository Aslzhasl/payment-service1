// internal/repository/deposit_repo.go
package repository

import (
	"context"
	"time"
)

// Deposit описывает запись из таблицы deposits
// Служит для хранения информации о депозите, связанном с бронью/листингом
type Deposit struct {
	StripePIID string    `db:"stripe_pi_id"`
	BookingID  string    `db:"booking_id"`
	ListingID  string    `db:"listing_id"`
	UserID     string    `db:"user_id"`
	Amount     int64     `db:"amount"`
	Currency   string    `db:"currency"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// DepositRepo описывает операции над таблицей deposits
// Хранилище (Store) должно реализовать эти методы
// для управления депозитами (holds) пользователей
type DepositRepo interface {
	// CreateDeposit сохраняет новый депозит (PaymentIntent) с ручным захватом
	CreateDeposit(ctx context.Context, d Deposit) error
	// UpdateDepositStatus обновляет статус существующего депозита
	UpdateDepositStatus(ctx context.Context, stripePIID, status string) error
	// GetDepositByID возвращает депозит по Stripe PaymentIntent ID
	GetDepositByID(ctx context.Context, stripePIID string) (Deposit, error)
	// ListDepositsByBookingID возвращает все депозиты, связанные с конкретной бронью
	ListDepositsByBookingID(ctx context.Context, bookingID string) ([]Deposit, error)
	// ListDepositsByUserID возвращает все депозиты данного пользователя
	ListDepositsByUserID(ctx context.Context, userID string) ([]Deposit, error)
}
