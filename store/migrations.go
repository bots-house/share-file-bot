package store

import "context"

// StoreMigrator define method for migrations of store
type StoreMigrator interface {
	Migrator() Migrator
}

// Migrator defines generic interface for migrations.
type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}
