package store

import "context"

// Migrator defines generic interface for migrations.
type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}

type TxFunc func(ctx context.Context) error

type Txier func(ctx context.Context, txFunc TxFunc) error

type Database interface {
	// Migrator returns database migration
	Migrator() Migrator

	// TxFactory returns function for create transaction scopes.
	Tx(ctx context.Context, txFunc TxFunc) error
}
