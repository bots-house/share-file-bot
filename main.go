package main

import (
	"context"
	"database/sql"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bots-house/share-file-bot/bot"
	"github.com/bots-house/share-file-bot/pkg/health"
	"github.com/bots-house/share-file-bot/service"
	"github.com/bots-house/share-file-bot/store/postgres"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/subosito/gotenv"
)

const (
	EnvLocal      = "local"
	EnvStaging    = "staging"
	EnvProduction = "production"
)

// Config represents service configuration
type Config struct {
	Env string `split_words:"true" default:"local"`

	SentryDSN string `split_words:"true"`

	Database             string `default:"postgres://sfb:sfb@localhost/sfb?sslmode=disable"`
	DatabaseMaxOpenConns int    `default:"10" split_words:"true"`
	DatabaseMaxIdleConns int    `default:"0" split_words:"true"`

	Token        string `required:"true"`
	Addr         string `default:":8000"`
	WebhookURL   string `default:"/" split_words:"true"`
	SecretIDSalt string `required:"true" split_words:"true"`

	DryRun bool `default:"false" split_words:"true"`

	LogDebug  bool `default:"true" split_words:"true"`
	LogPretty bool `default:"false" split_words:"true"`
}

func (cfg Config) getEnv() string {
	for _, v := range []string{EnvLocal, EnvProduction, EnvStaging} {
		print(v)
		if v == strings.ToLower(cfg.Env) {
			return v
		}
	}
	return EnvLocal
}

var revision = "unknown"

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

	// parse config
	var (
		cfg        Config
		flagConfig string
	)

	// parse config
	cfg, err := parseConfig(flagConfig)
	if err != nil {
		os.Exit(1)
	}

	ctx = setupLogging(ctx, cfg)

	if err := run(ctx, cfg); err != nil {
		log.Ctx(ctx).Error().Str("err", err.Error()).Msg("fatal error")
		cancel()
		//nolint: gocritic
		os.Exit(1)
	}
}

func newServer(addr string, bot *bot.Bot, db *sql.DB, cfg Config) *http.Server {
	baseCtx := context.Background()
	baseCtx = setupLogging(baseCtx, cfg)

	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	return &http.Server{
		Addr:    addr,
		Handler: sentryMiddleware.Handle(newMux(bot, db)),
		BaseContext: func(_ net.Listener) context.Context {
			return baseCtx
		},
	}
}

func newMux(bot *bot.Bot, db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/health", health.NewHandler(db))

	mux.Handle("/", bot)

	return mux
}

func newSentry(ctx context.Context, cfg Config, release string) error {
	env := cfg.getEnv()

	if env == EnvLocal {
		log.Ctx(ctx).Debug().Str("env", env).Msg("sentry is not available in this env")
		return nil
	}

	if cfg.SentryDSN == "" {
		log.Ctx(ctx).Warn().Str("env", env).Msg("sentry dsn is not provided")
		return nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Env,
		Release:     release,
	}); err != nil {
		return errors.Wrap(err, "init sentry")
	}

	return nil
}

const envPrefix = "SFB"

func run(ctx context.Context, cfg Config) error {
	// parse flags
	var (
		flagHealth bool
		flagConfig string
	)

	flag.BoolVar(&flagHealth, "health", false, "run health check")
	flag.StringVar(&flagConfig, "config", "", "load env from file")

	flag.Parse()

	if flagHealth {
		return health.Check(ctx, cfg.Addr)
	}

	log.Ctx(ctx).Info().Str("revision", revision).Msg("start")
	if err := newSentry(ctx, cfg, revision); err != nil {
		return errors.Wrap(err, "init sentry")
	}

	log.Ctx(ctx).Info().Str("dsn", cfg.Database).Msg("open db")

	// open and ping db
	db, err := sql.Open("postgres", cfg.Database)
	if err != nil {
		return errors.Wrap(err, "open db")
	}
	defer db.Close()

	log.Ctx(ctx).Debug().Msg("ping database")
	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "ping db")
	}

	db.SetMaxOpenConns(cfg.DatabaseMaxOpenConns)
	db.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)

	// create abstraction around db and apply migrations
	pg := postgres.NewPostgres(db)

	log.Ctx(ctx).Info().Msg("migrate database")
	if err := pg.Migrator().Up(ctx); err != nil {
		return errors.Wrap(err, "migrate db")
	}

	authSrv := &service.Auth{
		UserStore: pg.User,
	}

	fileSrv := &service.File{
		FileStore:     pg.File,
		DownloadStore: pg.Download,
	}

	adminSrv := &service.Admin{
		User:     pg.User,
		File:     pg.File,
		Download: pg.Download,
	}

	log.Ctx(ctx).Info().Msg("init bot")
	tgBot, err := bot.New(revision, cfg.Token, authSrv, fileSrv, adminSrv)
	if err != nil {
		return errors.Wrap(err, "init bot")
	}

	log.Ctx(ctx).Info().Str("link", "https://t.me/"+tgBot.Self().UserName).Msg("bot is alive")

	server := newServer(cfg.Addr, tgBot, db, cfg)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		log.Ctx(ctx).Info().Msg("shutdown server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Ctx(ctx).Warn().Str("err", err.Error()).Msg("shutdown error")
		}
	}()

	if err := tgBot.SetWebhookIfNeed(ctx, cfg.WebhookURL); err != nil {
		return errors.Wrap(err, "set webhook if need")
	}

	// if we run in dry run mode, exit without blocking
	if cfg.DryRun {
		return nil
	}

	log.Ctx(ctx).Info().Str("addr", cfg.Addr).Str("webhook_domain", cfg.WebhookURL).Msg("start server")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "listen and serve")
	}

	return nil
}

func parseConfig(config string) (Config, error) {
	var cfg Config

	// load envs
	if config != "" {
		if err := gotenv.Load(config); err != nil {
			return cfg, errors.Wrap(err, "load env")
		}
	}

	if err := envconfig.Process(envPrefix, &cfg); err != nil {
		_ = envconfig.Usage(envPrefix, &cfg)
		return cfg, err
	}

	return cfg, nil
}

func setupLogging(ctx context.Context, cfg Config) context.Context {
	if cfg.LogPretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if cfg.LogDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return log.Logger.WithContext(ctx)
}
