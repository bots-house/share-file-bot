package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/store"
	"github.com/bots-house/share-file-bot/store/postgres/migrations"
	"github.com/bots-house/share-file-bot/store/postgres/shared"

	"github.com/friendsofgo/errors"

	// import postgresq driver
	_ "github.com/lib/pq"
)

type Postgres struct {
	*sql.DB
	migrator *migrations.Migrator

	user     *UserStore
	file     *FileStore
	download *DownloadStore
	chat     *ChatStore
	bot      *BotStore
}

var _ store.Store = &Postgres{}

func (pg *Postgres) File() core.FileStore {
	return pg.file
}

func (pg *Postgres) User() core.UserStore {
	return pg.user
}

func (pg *Postgres) Download() core.DownloadStore {
	return pg.download
}

func (pg *Postgres) Chat() core.ChatStore {
	return pg.chat
}

func (pg *Postgres) Bot() core.BotStore {
	return pg.bot
}

// New create postgres based database with all stores.
func New(db *sql.DB) *Postgres {
	pg := &Postgres{
		DB:       db,
		migrator: migrations.New(db),
	}

	base := BaseStore{
		DB:    db,
		Txier: pg.Tx,
	}

	pg.download = &DownloadStore{base}
	pg.user = &UserStore{base}
	pg.file = &FileStore{base}
	pg.chat = &ChatStore{base}
	pg.bot = &BotStore{base}

	return pg
}

func (pg *Postgres) Migrator() store.Migrator {
	return pg.migrator
}

// Tx run code in database transaction.
// Based on: https://stackoverflow.com/a/23502629.
func (pg *Postgres) Tx(ctx context.Context, txFunc store.TxFunc) (err error) {
	tx := shared.GetTx(ctx)

	if tx != nil {
		return txFunc(ctx)
	}

	tx, err = pg.BeginTx(ctx, nil)
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
