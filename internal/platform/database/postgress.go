package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver
)

// NewPostgres creates a connection pool with production settings
func NewPostgres(uri string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", uri)
	if err != nil {
		return nil, err
	}

	// Production settings: prevent connection exhaustion
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// ExecTx provides a helper to wrap logic in a transaction
func ExecTx(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
