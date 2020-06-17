package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/store/postgres"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	Database string `default:"postgres://sfb:sfb@localhost/sfb?sslmode=disable"`
}

func main() {
	logger := log.NewLogger(true, true)

	ctx := context.Background()
	ctx = log.WithLogger(ctx, logger)
	if err := run(ctx); err != nil {
		log.Error(ctx, "fatal error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// parse config
	var cfg Config

	if err := envconfig.Process("SFB", &cfg); err != nil {
		return errors.Wrap(err, "parse config from env")
	}

	log.Info(ctx, "open db", "dsn", cfg.Database)

	// open and ping db
	db, err := sql.Open("postgres", cfg.Database)
	if err != nil {
		return errors.Wrap(err, "open db")
	}

	log.Debug(ctx, "ping database")
	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "ping db")
	}

	// create abstraction around db and apply migrations
	pg := postgres.NewPostgres(db)

	log.Info(ctx, "migrate database")
	if err := pg.Migrator().Up(ctx); err != nil {
		return errors.Wrap(err, "migrate db")
	}

	return nil
}
