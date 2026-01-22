package account

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for accounts
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new account repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetAccountWithLock uses SELECT FOR UPDATE to lock the row
// This is critical to prevent race conditions in balance updates
func (r *Repository) GetAccountWithLock(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error) {
	var acc Account
	query := `SELECT id, balance_cents, currency, created_at, updated_at FROM accounts WHERE id = $1 FOR UPDATE`

	err := tx.GetContext(ctx, &acc, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account %s not found", id)
		}
		return nil, fmt.Errorf("failed to lock account %s: %w", id, err)
	}
	return &acc, nil
}

// GetAccount retrieves an account without locking
func (r *Repository) GetAccount(ctx context.Context, id string) (*Account, error) {
	var acc Account
	query := `SELECT id, balance_cents, currency, created_at, updated_at FROM accounts WHERE id = $1`

	err := r.db.GetContext(ctx, &acc, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account %s not found", id)
		}
		return nil, fmt.Errorf("failed to get account %s: %w", id, err)
	}
	return &acc, nil
}

// UpdateBalance updates the balance of an account within a transaction
func (r *Repository) UpdateBalance(ctx context.Context, tx *sqlx.Tx, id string, amount int64) error {
	query := `UPDATE accounts SET balance_cents = balance_cents + $1, updated_at = NOW() WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to update balance for account %s: %w", id, err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("account %s not found", id)
	}
	
	return nil
}

// CreateAccount creates a new account
func (r *Repository) CreateAccount(ctx context.Context, acc *Account) error {
	query := `INSERT INTO accounts (id, balance_cents, currency, created_at, updated_at) 
	          VALUES ($1, $2, $3, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, acc.ID, acc.BalanceCents, acc.Currency)
	if err != nil {
		return fmt.Errorf("failed to create account %s: %w", acc.ID, err)
	}
	return nil
}

// UpdateAccount updates account currency
func (r *Repository) UpdateAccount(ctx context.Context, id string, currency string) error {
	query := `UPDATE accounts SET currency = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, currency, id)
	if err != nil {
		return fmt.Errorf("failed to update account %s: %w", id, err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("account %s not found", id)
	}
	
	return nil
}

// DeleteAccount deletes an account
func (r *Repository) DeleteAccount(ctx context.Context, id string) error {
	query := `DELETE FROM accounts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account %s: %w", id, err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("account %s not found", id)
	}
	
	return nil
}

// GetAllAccounts retrieves all accounts with pagination
func (r *Repository) GetAllAccounts(ctx context.Context, limit, offset int) ([]Account, error) {
	var accounts []Account
	query := `SELECT id, balance_cents, currency, created_at, updated_at FROM accounts ORDER BY id LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &accounts, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return accounts, nil
}

// GetAccountCount returns total number of accounts
func (r *Repository) GetAccountCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM accounts`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to get account count: %w", err)
	}
	return count, nil
}