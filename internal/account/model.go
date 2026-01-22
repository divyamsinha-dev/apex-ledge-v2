package account

import "time"

// Account represents the database entity
type Account struct {
	ID           string    `db:"id"`
	BalanceCents int64     `db:"balance_cents"`
	Currency     string    `db:"currency"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// TransferEvent is used for the async worker pool
type TransferEvent struct {
	FromID string
	ToID   string
	Amount int64
}
