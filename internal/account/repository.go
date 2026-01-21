package ledger

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository uses Generics to handle any model type
type Repository[T any] struct {
	db *sqlx.DB
}

func NewRepository[T any](db *sqlx.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// GetAccountWithLock uses SELECT FOR UPDATE to lock the row
// This is critical to prevent race conditions in balance updates
func (r *Repository[T]) GetAccountWithLock(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error) {
	var acc Account
	query := `SELECT id, balance_cents, currency FROM accounts WHERE id = $1 FOR UPDATE`

	err := tx.GetContext(ctx, &acc, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to lock account %s: %w", id, err)
	}
	return &acc, nil
}
