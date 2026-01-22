package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// NewPostgres creates a connection pool with production settings using pgx/v5
func NewPostgres(uri string) (*sqlx.DB, error) {
	// Parse the connection string
	config, err := pgx.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	// Use pgx/v5 stdlib driver with sqlx
	db := stdlib.OpenDB(*config)

	// Wrap with sqlx
	sqlxDB := sqlx.NewDb(db, "pgx")

	// Production settings: prevent connection exhaustion
	sqlxDB.SetMaxOpenConns(25)
	sqlxDB.SetMaxIdleConns(25)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := sqlxDB.Ping(); err != nil {
		return nil, err
	}

	return sqlxDB, nil
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
