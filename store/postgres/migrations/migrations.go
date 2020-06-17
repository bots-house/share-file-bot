package migrations

import (
	"context"
	"database/sql"

	"github.com/remind101/migrate"
)

// Migrator execute migrations.
type Migrator struct {
	migrator *migrate.Migrator
}

// New create a migrator.
func New(db *sql.DB) *Migrator {
	return &Migrator{
		migrator: migrate.NewPostgresMigrator(db),
	}
}

// Up database migrations
func (m *Migrator) Up(ctx context.Context) error {
	return m.migrator.Exec(migrate.Up, migrations...)
}

// Down database migrations
func (m *Migrator) Down(ctx context.Context) error {
	return m.migrator.Exec(migrate.Down, migrations...)
}

func query(queries ...string) func(*sql.Tx) error {
	return func(tx *sql.Tx) error {
		for _, query := range queries {
			if _, err := tx.Exec(query); err != nil {
				return err
			}
		}

		return nil
	}
}

var migrations []migrate.Migration

// include migration to list
func include(id int, up, down func(tx *sql.Tx) error) {
	migrations = append(migrations, migrate.Migration{
		ID:   id,
		Up:   up,
		Down: down,
	})
}
