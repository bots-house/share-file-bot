package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bots-house/share-file-bot/bot"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/secretid"
	"github.com/bots-house/share-file-bot/service"
	"github.com/bots-house/share-file-bot/store/postgres"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	Database             string `default:"postgres://sfb:sfb@localhost/sfb?sslmode=disable"`
	DatabaseMaxOpenConns int    `default:"10" split_words:"true"`
	DatabaseMaxIdleConns int    `default:"0" split_words:"true"`

	Token         string `required:"true"`
	Addr          string `default:":8000"`
	WebhookDomain string `required:"true" split_words:"true"`
	WebhookPath   string `default:"/" split_words:"true"`
	SecretIDSalt  string `required:"true" split_words:"true"`
}

var logger = log.NewLogger(true, true)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)

		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		<-sig

		cancel()
	}()

	ctx = log.WithLogger(ctx, logger)
	if err := run(ctx); err != nil {
		log.Error(ctx, "fatal error", "err", err)
		os.Exit(1)
	}
}

func newServer(addr string, bot *bot.Bot) *http.Server {
	baseCtx := context.Background()
	baseCtx = log.WithLogger(baseCtx, logger)

	return &http.Server{
		Addr:    addr,
		Handler: bot,
		BaseContext: func(_ net.Listener) context.Context {
			return baseCtx
		},
	}
}

const envPrefix = "SFB"

func run(ctx context.Context) error {
	// parse config
	var cfg Config

	if err := envconfig.Process(envPrefix, &cfg); err != nil {
		envconfig.Usage(envPrefix, &cfg)
		return errors.Wrap(err, "parse config from env")
	}

	log.Info(ctx, "open db", "dsn", cfg.Database)

	// open and ping db
	db, err := sql.Open("postgres", cfg.Database)
	if err != nil {
		return errors.Wrap(err, "open db")
	}
	defer db.Close()

	log.Debug(ctx, "ping database")
	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "ping db")
	}

	db.SetMaxOpenConns(cfg.DatabaseMaxOpenConns)
	db.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)

	// create abstraction around db and apply migrations
	pg := postgres.NewPostgres(db)

	log.Info(ctx, "migrate database")
	if err := pg.Migrator().Up(ctx); err != nil {
		return errors.Wrap(err, "migrate db")
	}

	authSrv := &service.Auth{
		UserStore: pg.User,
	}

	secretID, err := secretid.NewHashIDs(cfg.SecretIDSalt)
	if err != nil {
		return errors.Wrap(err, "init secret id")
	}

	docSrv := &service.Document{
		SecretID:      secretID,
		DocumentStore: pg.Document,
		DownloadStore: pg.Download,
	}

	log.Info(ctx, "init bot")
	tgBot, err := bot.New(cfg.Token, authSrv, docSrv)
	if err != nil {
		return errors.Wrap(err, "init bot")
	}

	server := newServer(cfg.Addr, tgBot)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		log.Info(ctx, "shutdown server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Warn(ctx, "shutdown error", "err", err)
		}
	}()

	if err := tgBot.SetWebhookIfNeed(ctx, cfg.WebhookDomain, cfg.WebhookPath); err != nil {
		return errors.Wrap(err, "set webhook if need")
	}

	log.Info(ctx, "start server", "addr", cfg.Addr, "webhook_domain", cfg.WebhookDomain)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "listen and serve")
	}

	return nil
}
