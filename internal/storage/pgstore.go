package storage

import (
	"Payment-service/internal/repository"
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Store оборачивает sqlx.DB и реализует репозитории.
type Store struct {
	DB *sqlx.DB
}

// InitStore открывает соединение по DATABASE_URL и возвращает Store.
func InitStore(databaseURL string) (*Store, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}

// Close закрывает соединение к базе.
func (s *Store) Close() error {
	return s.DB.Close()
}

// CreateCustomer создаёт запись пользователя в таблице customers.
func (s *Store) CreateCustomer(ctx context.Context, userID, email, stripeID string) error {
	query := `
    INSERT INTO customers (user_id, email, stripe_id, created_at)
    VALUES ($1, $2, $3, now())
    ON CONFLICT (user_id) DO NOTHING;
    `
	_, err := s.DB.ExecContext(ctx, query, userID, email, stripeID)
	return err
}

// GetCustomerByUserID возвращает Stripe Customer ID для заданного user_id.
func (s *Store) GetCustomerByUserID(ctx context.Context, userID string) (string, error) {
	var stripeID string
	query := `SELECT stripe_id FROM customers WHERE user_id = $1`
	err := s.DB.GetContext(ctx, &stripeID, query, userID)
	return stripeID, err
}

// SavePaymentMethod сохраняет новую карту в таблице payment_methods.
func (s *Store) SavePaymentMethod(ctx context.Context, pm repository.PaymentMethod) error {
	query := `
    INSERT INTO payment_methods
      (user_id, stripe_pm_id, card_brand, card_last4, exp_month, exp_year, created_at)
    VALUES ($1, $2, $3, $4, $5, $6, now())
    ON CONFLICT (stripe_pm_id) DO NOTHING;
    `
	_, err := s.DB.ExecContext(ctx, query,
		pm.UserID, pm.StripePMID, pm.Brand, pm.Last4, pm.ExpMonth, pm.ExpYear,
	)
	return err
}

// ListPaymentMethods возвращает все карты пользователя.
func (s *Store) ListPaymentMethods(ctx context.Context, userID string) ([]repository.PaymentMethod, error) {
	var methods []repository.PaymentMethod
	query := `
    SELECT user_id, stripe_pm_id, card_brand, card_last4, exp_month, exp_year, created_at
    FROM payment_methods
    WHERE user_id = $1
    ORDER BY created_at DESC;
    `
	err := s.DB.SelectContext(ctx, &methods, query, userID)
	return methods, err
}

// CreatePaymentIntent сохраняет новый PaymentIntent в таблице payment_intents.
func (s *Store) CreatePaymentIntent(ctx context.Context, pi repository.PaymentIntent) error {
	query := `
    INSERT INTO payment_intents
      (stripe_pi_id, booking_id, user_id, amount, currency, status, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, now(), now())
    ON CONFLICT (stripe_pi_id) DO NOTHING;
    `
	_, err := s.DB.ExecContext(ctx, query,
		pi.StripePIID, pi.BookingID, pi.UserID, pi.Amount, pi.Currency, pi.Status,
	)
	return err
}

// UpdatePaymentIntentStatus обновляет статус существующего платежа.
func (s *Store) UpdatePaymentIntentStatus(ctx context.Context, stripePIID, status string) error {
	query := `
    UPDATE payment_intents
    SET status = $2, updated_at = now()
    WHERE stripe_pi_id = $1;
    `
	_, err := s.DB.ExecContext(ctx, query, stripePIID, status)
	return err
}

// GetPaymentIntentByID возвращает запись PaymentIntent по его Stripe ID.
func (s *Store) GetPaymentIntentByID(ctx context.Context, stripePIID string) (repository.PaymentIntent, error) {
	var pi repository.PaymentIntent
	query := `
    SELECT stripe_pi_id, booking_id, user_id, amount, currency, status, created_at, updated_at
    FROM payment_intents
    WHERE stripe_pi_id = $1;
    `
	err := s.DB.GetContext(ctx, &pi, query, stripePIID)
	return pi, err
}

// --- Реализация интерфейсов репозиториев ---A

var (
	_ repository.CustomerRepo      = (*Store)(nil)
	_ repository.PaymentMethodRepo = (*Store)(nil)
	_ repository.PaymentIntentRepo = (*Store)(nil)
)

// CustomerRepo
func (s *Store) CreateCustomerRecord(ctx context.Context, userID, email, stripeID string) error {
	return s.CreateCustomer(ctx, userID, email, stripeID)
}

func (s *Store) GetCustomerByUser(ctx context.Context, userID string) (string, error) {
	return s.GetCustomerByUserID(ctx, userID)
}

// PaymentMethodRepo
func (s *Store) SavePaymentMethodRecord(ctx context.Context, pm repository.PaymentMethod) error {
	return s.SavePaymentMethod(ctx, pm)
}

func (s *Store) ListPaymentMethodsByUser(ctx context.Context, userID string) ([]repository.PaymentMethod, error) {
	return s.ListPaymentMethods(ctx, userID)
}

// PaymentIntentRepo
func (s *Store) CreatePaymentIntentRecord(ctx context.Context, pi repository.PaymentIntent) error {
	return s.CreatePaymentIntent(ctx, pi)
}

func (s *Store) UpdatePaymentIntentStatusRecord(ctx context.Context, stripePIID, status string) error {
	return s.UpdatePaymentIntentStatus(ctx, stripePIID, status)
}

func (s *Store) GetPaymentIntentByIDRecord(ctx context.Context, stripePIID string) (repository.PaymentIntent, error) {
	return s.GetPaymentIntentByID(ctx, stripePIID)
}

// --- DepositRepo ---

// CreateDeposit сохраняет новый депозит в таблице deposits.
func (s *Store) CreateDeposit(ctx context.Context, d repository.Deposit) error {
	const query = `
INSERT INTO deposits
  (stripe_pi_id, booking_id, listing_id, user_id, amount, currency, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (stripe_pi_id) DO NOTHING;
`
	_, err := s.DB.ExecContext(ctx, query,
		d.StripePIID, d.BookingID, d.ListingID, d.UserID,
		d.Amount, d.Currency, d.Status,
	)
	return err
}

// UpdateDepositStatus обновляет статус депозита.
func (s *Store) UpdateDepositStatus(ctx context.Context, stripePIID, status string) error {
	const query = `
UPDATE deposits
SET status = $2, updated_at = now()
WHERE stripe_pi_id = $1;
`
	_, err := s.DB.ExecContext(ctx, query, stripePIID, status)
	return err
}

// GetDepositByID возвращает депозит по stripe_pi_id.
func (s *Store) GetDepositByID(ctx context.Context, stripePIID string) (repository.Deposit, error) {
	const query = `
SELECT stripe_pi_id, booking_id, listing_id, user_id, amount, currency, status, created_at, updated_at
FROM deposits
WHERE stripe_pi_id = $1;
`
	var d repository.Deposit
	err := s.DB.GetContext(ctx, &d, query, stripePIID)
	return d, err
}

// ListDepositsByBookingID возвращает депозиты для конкретной брони.
func (s *Store) ListDepositsByBookingID(ctx context.Context, bookingID string) ([]repository.Deposit, error) {
	const query = `
SELECT stripe_pi_id, booking_id, listing_id, user_id, amount, currency, status, created_at, updated_at
FROM deposits
WHERE booking_id = $1
ORDER BY created_at DESC;
`
	var list []repository.Deposit
	err := s.DB.SelectContext(ctx, &list, query, bookingID)
	return list, err
}

// ListDepositsByUserID возвращает депозиты для пользователя.
func (s *Store) ListDepositsByUserID(ctx context.Context, userID string) ([]repository.Deposit, error) {
	const query = `
SELECT stripe_pi_id, booking_id, listing_id, user_id, amount, currency, status, created_at, updated_at
FROM deposits
WHERE user_id = $1
ORDER BY created_at DESC;
`
	var list []repository.Deposit
	err := s.DB.SelectContext(ctx, &list, query, userID)
	return list, err
}

// Проверка, что Store реализует DepositRepo
var _ repository.DepositRepo = (*Store)(nil)
