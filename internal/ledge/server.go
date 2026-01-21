package ledger

import (
	"context"
	"fmt"
	"sync"
)

type Account struct {
	ID           string `db:"id"`
	BalanceCents int64  `db:"balance_cents"`
	Currency     string `db:"currency"`
}

type Server struct {
	repo *Repository[Account]
	mu   sync.RWMutex // Protects in-memory state if needed
}

func (s *Server) Transfer(ctx context.Context, fromID, toID string, amount int64) error {
	// 1. Start a SQL Transaction via the repo's DB handle
	tx, err := s.repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Safety: Rollback if we don't commit

	// 2. Lock both accounts (Pessimistic Locking)
	// TIP: Always lock accounts in a sorted order (e.g., by ID) to avoid Deadlocks!
	fromAcc, err := s.repo.GetAccountWithLock(ctx, tx, fromID)
	if err != nil {
		return err
	}

	toAcc, err := s.repo.GetAccountWithLock(ctx, tx, toID)
	if err != nil {
		return err
	}

	// 3. Business Logic: Check for sufficient funds
	if fromAcc.BalanceCents < amount {
		return fmt.Errorf("insufficient funds in account %s", fromID)
	}

	// 4. Perform Double-Entry Updates
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2", amount, fromID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance_cents = balance_cents + $1 WHERE id = $2", amount, toID)
	if err != nil {
		return err
	}

	// 5. Commit the transaction
	return tx.Commit()
}
