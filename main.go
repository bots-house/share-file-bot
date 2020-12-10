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
	"github.com/bots-house/share-file-bot/bot/state"
	"github.com/bots-house/share-file-bot/pkg"
	"github.com/bots-house/share-file-bot/pkg/health"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/service"
	"github.com/bots-house/share-file-bot/store/postgres"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	redis "github.com/go-redis/redis/v8"

	"github.com/kelseyhightower/envconfig"
	"github.com/subosito/gotenv"
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

	Redis             string `default:"redis://localhost:6379"`
	RedisMaxOpenConns int    `default:"10" split_words:"true"`
	RedisMaxIdleConns int    `default:"0" split_words:"true"`

	Token        string `required:"true"`
	Addr         string `default:":8000"`
	WebhookURL   string `default:"/" split_words:"true"`
	SecretIDSalt string `required:"true" split_words:"true"`

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

var (
	buildVersion = "unknown"
	buildRef     = "unknown"
	buildTime    = "unknown"
)

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
		cancel()
		//nolint: gocritic
		os.Exit(1)
	}
}

func newServer(addr string, bot *bot.Bot, db *sql.DB) *http.Server {
	baseCtx := context.Background()
	baseCtx = log.WithLogger(baseCtx, logger)

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
		log.Debug(ctx, "sentry is not available in this env", "env", env)
		return nil
	}

	if cfg.SentryDSN == "" {
		log.Warn(ctx, "sentry dsn is not provided", "env", env)
		return nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Env,
		Release:          release,
		AttachStacktrace: true,
	}); err != nil {
		return errors.Wrap(err, "init sentry")
	}

	return nil
}

const envPrefix = "SFB"

func run(ctx context.Context) error {
	buildInfo := pkg.BuildInfo{
		Version: buildVersion,
		Ref:     buildRef,
		Time:    buildTime,
	}

	// parse config
	var cfg Config

	// parse flags
	var (
		flagHealth bool
		flagConfig string
	)

	flag.BoolVar(&flagHealth, "health", false, "run health check")
	flag.StringVar(&flagConfig, "config", "", "load env from file")

	flag.Parse()

	// parse config
	cfg, err := parseConfig(flagConfig)
	if err != nil {
		return errors.Wrap(err, "parse config")
	}

	if flagHealth {
		return health.Check(ctx, cfg.Addr)
	}

	log.Info(ctx, "start",
		"version", buildInfo.Version,
		"ref", buildInfo.Ref,
		"time", buildInfo.Time,
	)

	if err := newSentry(ctx, cfg, buildInfo.Version); err != nil {
		return errors.Wrap(err, "init sentry")
	}
	defer sentry.Flush(time.Second * 5)

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
	pg := postgres.New(db)

	log.Info(ctx, "migrate database")
	if err := pg.Migrator().Up(ctx); err != nil {
		return errors.Wrap(err, "migrate db")
	}

	log.Info(ctx, "open redis",
		"dsn", cfg.Redis,
		"max_open_conns",
		cfg.RedisMaxOpenConns,
		"max_idle_conns",
		cfg.RedisMaxIdleConns,
	)

	rdbOpts, err := redis.ParseURL(cfg.Redis)
	if err != nil {
		return errors.Wrap(err, "parse redis url")
	}

	rdb := redis.NewClient(rdbOpts)

	log.Debug(ctx, "ping redis")
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return errors.Wrap(err, "ping redis")
	}

	botState := state.NewRedisStore(rdb, "share-file-bot")

	log.Info(ctx, "init bot api client")
	tgClient, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return errors.Wrap(err, "create bot api")
	}

	authSrv := &service.Auth{
		UserStore: pg.User(),
	}

	fileSrv := &service.File{
		File:     pg.File(),
		Chat:     pg.Chat(),
		Download: pg.Download(),
		Telegram: tgClient,
		Redis:    rdb,
	}

	adminSrv := &service.Admin{
		User:     pg.User(),
		File:     pg.File(),
		Download: pg.Download(),
		Chat:     pg.Chat(),
	}

	chatSrv := &service.Chat{
		Telegram: tgClient,
		Txier:    pg.Tx,
		Chat:     pg.Chat(),
		File:     pg.File(),
		Download: pg.Download(),
	}

	tgBot, err := bot.New(buildInfo, tgClient, botState, authSrv, fileSrv, adminSrv, chatSrv)
	if err != nil {
		return errors.Wrap(err, "init bot")
	}
	log.Info(ctx, "bot is alive", "link", "https://t.me/"+tgBot.Self().UserName)

	server := newServer(cfg.Addr, tgBot, db)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		log.Info(ctx, "shutdown server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Warn(ctx, "shutdown error", "err", err)
		}
	}()

	if err := tgBot.SetWebhookIfNeed(ctx, cfg.WebhookURL); err != nil {
		return errors.Wrap(err, "set webhook if need")
	}

	// if we run in dry run mode, exit without blocking
	if cfg.DryRun {
		return nil
	}

	log.Info(ctx, "start server", "addr", cfg.Addr, "webhook_domain", cfg.WebhookURL)
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
