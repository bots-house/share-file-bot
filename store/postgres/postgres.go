package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/store"
	"github.com/bots-house/share-file-bot/store/postgres/migrations"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Postgres struct {
	*sql.DB
	migrator *migrations.Migrator

	User     *UserStore
	File     *FileStore
	Download *DownloadStore
}

// NewPostgres create postgres based database with all stores.
func NewPostgres(db *sql.DB) *Postgres {
	pg := &Postgres{
		DB:       db,
		migrator: migrations.New(db),
		User:     &UserStore{ContextExecutor: db},
		File:     &FileStore{ContextExecutor: db},
		Download: &DownloadStore{ContextExecutor: db},
	}

	return pg
}

func (p *Postgres) Migrator() store.Migrator {
	return p.migrator
}

// Tx run code in database transaction.
// Based on: https://stackoverflow.com/a/23502629.
func (p *Postgres) Tx(ctx context.Context, txFunc store.TxFunc) (err error) {
	tx := shared.GetTx(ctx)

	if tx != nil {
		return txFunc(ctx)
	}

	tx, err = p.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx failed")
	}

	ctx = shared.WithTx(ctx, tx)

	//nolint:gocritic
	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback(); err != nil {
				log.Warn(ctx, "tx rollback failed", "err", err)
			}
			panic(r)
		} else if err != nil {
			if err := tx.Rollback(); err != nil {
				log.Warn(ctx, "tx rollback failed", "err", err)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = txFunc(ctx)

	return err
}
