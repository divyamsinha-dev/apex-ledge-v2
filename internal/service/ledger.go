package service

import (
	"context"
	"fmt"
	"time"

	"apex-ledger/internal/account"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// LedgerService handles business logic for ledger operations
type LedgerService struct {
	accountRepo *account.Repository
	db          *sqlx.DB
}

// NewLedgerService creates a new ledger service
func NewLedgerService(accountRepo *account.Repository, db *sqlx.DB) *LedgerService {
	return &LedgerService{
		accountRepo: accountRepo,
		db:          db,
	}
}

// PerformTransfer executes a double-entry transfer between two accounts
func (s *LedgerService) PerformTransfer(ctx context.Context, fromID, toID string, amount int64) (string, error) {
	// Validate inputs
	if fromID == "" || toID == "" {
		return "", fmt.Errorf("account IDs cannot be empty")
	}
	if fromID == toID {
		return "", fmt.Errorf("cannot transfer to the same account")
	}
	if amount <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}

	// Generate transaction ID
	txID := uuid.New().String()

	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock both accounts in sorted order to prevent deadlocks
	// Always lock in alphabetical order
	var fromAcc, toAcc *account.Account
	if fromID < toID {
		fromAcc, err = s.accountRepo.GetAccountWithLock(ctx, tx, fromID)
		if err != nil {
			return "", err
		}
		toAcc, err = s.accountRepo.GetAccountWithLock(ctx, tx, toID)
		if err != nil {
			return "", err
		}
	} else {
		toAcc, err = s.accountRepo.GetAccountWithLock(ctx, tx, toID)
		if err != nil {
			return "", err
		}
		fromAcc, err = s.accountRepo.GetAccountWithLock(ctx, tx, fromID)
		if err != nil {
			return "", err
		}
	}

	// Check currency match
	if fromAcc.Currency != toAcc.Currency {
		return "", fmt.Errorf("currency mismatch: %s != %s", fromAcc.Currency, toAcc.Currency)
	}

	// Check sufficient funds
	if fromAcc.BalanceCents < amount {
		return "", fmt.Errorf("insufficient funds in account %s: balance %d, required %d", fromID, fromAcc.BalanceCents, amount)
	}

	// Perform double-entry updates
	if err := s.accountRepo.UpdateBalance(ctx, tx, fromID, -amount); err != nil {
		return "", fmt.Errorf("failed to debit account %s: %w", fromID, err)
	}

	if err := s.accountRepo.UpdateBalance(ctx, tx, toID, amount); err != nil {
		return "", fmt.Errorf("failed to credit account %s: %w", toID, err)
	}

	// Record transaction in ledger (optional but recommended)
	if err := s.recordTransaction(ctx, tx, txID, fromID, toID, amount, fromAcc.Currency); err != nil {
		return "", fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return txID, nil
}

// GetBalance retrieves the current balance of an account
func (s *LedgerService) GetBalance(ctx context.Context, accountID string) (*account.Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	acc, err := s.accountRepo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// recordTransaction records the transfer in the transactions table
func (s *LedgerService) recordTransaction(ctx context.Context, tx *sqlx.Tx, txID, fromID, toID string, amount int64, currency string) error {
	query := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount_cents, currency, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.ExecContext(ctx, query, txID, fromID, toID, amount, currency, time.Now())
	return err
}
