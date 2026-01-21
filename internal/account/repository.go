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
	query := `SELECT id, balance_cents, currency, updated_at FROM accounts WHERE id = $1 FOR UPDATE`

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
	query := `SELECT id, balance_cents, currency, updated_at FROM accounts WHERE id = $1`

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
