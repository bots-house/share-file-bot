package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bots-house/share-file-bot/bot"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/service"
	"github.com/bots-house/share-file-bot/store/postgres"
	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	EnvLocal      = "local"
	EnvStaging    = "staging"
	EnvProduction = "production"
)

type Config struct {
	Env string `split_words:"true" default:"local"`

	SentryDSN string `split_words:"true"`

	Database             string `default:"postgres://sfb:sfb@localhost/sfb?sslmode=disable"`
	DatabaseMaxOpenConns int    `default:"10" split_words:"true"`
	DatabaseMaxIdleConns int    `default:"0" split_words:"true"`

	Token         string `required:"true"`
	Addr          string `default:":8000"`
	WebhookDomain string `required:"true" split_words:"true"`
	WebhookPath   string `default:"/" split_words:"true"`
	SecretIDSalt  string `required:"true" split_words:"true"`

	DryRun bool `default:"false" split_words:"true"`
}

func (cfg Config) getEnv() string {
	for _, v := range []string{EnvLocal, EnvProduction, EnvStaging} {
		if v == strings.ToLower(cfg.Env) {
			return v
		}
	}
	return EnvLocal
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

func newSentry(ctx context.Context, cfg Config) error {
	env := cfg.getEnv()

	if env == EnvLocal {
		log.Debug(ctx, "sentry is not available in this env", "env", env)
		return nil
	}

	if cfg.SentryDSN == "" {
		log.Warn(ctx, "sentry dsn is not provided", "env", env)
		return nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Env,
	}); err != nil {
		return errors.Wrap(err, "init sentry")
	}

	return nil
}

const envPrefix = "SFB"

func run(ctx context.Context) error {
	// parse config
	var cfg Config

	if err := envconfig.Process(envPrefix, &cfg); err != nil {
		_ = envconfig.Usage(envPrefix, &cfg)
		return errors.Wrap(err, "parse config from env")
	}

	if err := newSentry(ctx, cfg); err != nil {
		return errors.Wrap(err, "init sentry")
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

	docSrv := &service.Document{
		DocumentStore: pg.Document,
		DownloadStore: pg.Download,
	}

	adminSrv := &service.Admin{
		User:     pg.User,
		Document: pg.Document,
		Download: pg.Download,
	}

	log.Info(ctx, "init bot")
	tgBot, err := bot.New(cfg.Token, authSrv, docSrv, adminSrv)
	if err != nil {
		return errors.Wrap(err, "init bot")
	}
	log.Info(ctx, "bot", "username", tgBot.Self().UserName)

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

	// if we run in dry run mode, exit without blocking
	if cfg.DryRun {
		return nil
	}

	log.Info(ctx, "start server", "addr", cfg.Addr, "webhook_domain", cfg.WebhookDomain)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "listen and serve")
	}

	return nil
}
