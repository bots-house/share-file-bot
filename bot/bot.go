package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
	"github.com/fatih/structs"
	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"github.com/tomasen/realip"
)

type Bot struct {
	client *tgbotapi.BotAPI

	authSrv  *service.Auth
	docSrv   *service.Document
	adminSrv *service.Admin

	handler tg.Handler
}

func (bot *Bot) Self() tgbotapi.User {
	return bot.client.Self
}

func New(token string, authSrv *service.Auth, docSrv *service.Document, adminSrv *service.Admin) (*Bot, error) {
	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "create bot api")
	}

	bot := &Bot{
		client:   client,
		authSrv:  authSrv,
		docSrv:   docSrv,
		adminSrv: adminSrv,
	}

	// bot.client.Debug = true

	bot.initHandler()

	return bot, nil
}

func (bot *Bot) SetWebhookIfNeed(ctx context.Context, u string) error {
	webhook, err := bot.client.GetWebhookInfo()
	if err != nil {
		return errors.Wrap(err, "get webhook info")
	}

	if webhook.URL != u {
		u, err := url.Parse(u)
		if err != nil {
			return errors.Wrap(err, "invalid provided webhook url")
		}

		log.Info(ctx, "update bot webhook", "old", webhook.URL, "new", u.String())
		if _, err := bot.client.SetWebhook(tgbotapi.WebhookConfig{
			URL:            u,
			MaxConnections: 40,
		}); err != nil {
			return errors.Wrap(err, "update webhook")
		}
	}

	return nil
}

func (bot *Bot) initHandler() {
	authMiddleware := newAuthMiddleware(bot.authSrv)

	handler := authMiddleware(tg.HandlerFunc(bot.onUpdate))

	bot.handler = handler
}

var (
	cbqDocumentRefresh       = regexp.MustCompile(`^document:(\d+):refresh$`)
	cbqDocumentDelete        = regexp.MustCompile(`^document:(\d+):delete$`)
	cbqDocumentDeleteConfirm = regexp.MustCompile(`^document:(\d+):delete:confirm$`)
	cbqSettingsToggleLongIDs = regexp.MustCompile(`^` + callbackSettingsLongIDs + `$`)
)

func (bot *Bot) onUpdate(ctx context.Context, update *tgbotapi.Update) error {
	// spew.Dump(update)

	// handle message
	if msg := update.Message; msg != nil {

		// handle command
		switch msg.Command() {
		case "start":
			return bot.onStart(ctx, msg)
		case "help":
			return bot.onHelp(ctx, msg)
		case "admin":
			return bot.onAdmin(ctx, msg)
		case "settings":
			return bot.onSettings(ctx, msg)
		}

		// handle other
		if msg.Document != nil {
			return bot.onDocument(ctx, msg)
		} else {
			return bot.onNotDocument(ctx, msg)
		}
	}

	// handle callback queries
	if cbq := update.CallbackQuery; cbq != nil {
		data := cbq.Data
		switch {
		case len(cbqDocumentRefresh.FindStringIndex(data)) > 0:
			result := cbqDocumentRefresh.FindStringSubmatch(data)

			id, err := strconv.Atoi(result[1])
			if err != nil {
				return errors.Wrap(err, "parse cbq data")
			}

			return bot.onDocumentRefreshCBQ(ctx, cbq, id)

		case len(cbqDocumentDelete.FindStringIndex(data)) > 0:
			result := cbqDocumentDelete.FindStringSubmatch(data)

			id, err := strconv.Atoi(result[1])
			if err != nil {
				return errors.Wrap(err, "parse cbq data")
			}

			return bot.onDocumentDeleteCBQ(ctx, cbq, id)
		case len(cbqDocumentDeleteConfirm.FindStringIndex(data)) > 0:
			result := cbqDocumentDeleteConfirm.FindStringSubmatch(data)

			id, err := strconv.Atoi(result[1])
			if err != nil {
				return errors.Wrap(err, "parse cbq data")
			}

			return bot.onDocumentDeleteConfirmCBQ(ctx, cbq, id)
		case len(cbqSettingsToggleLongIDs.FindStringIndex(data)) > 0:
			return bot.onSettingsToggleLongIDsCBQ(ctx, cbq)
		}
	}

	return nil
}

func (bot *Bot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := realip.FromRequest(r)

	if !isTelegramIP(ip) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	update := &tgbotapi.Update{}

	// parse update
	if err := json.NewDecoder(r.Body).Decode(update); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}

	// handle update
	if err := bot.handler.HandleUpdate(ctx, update); err != nil {
		bot.onError(ctx, update, err)
		return
	}
}

func (bot *Bot) onError(ctx context.Context, update *tgbotapi.Update, er error) {
	log.Error(ctx, "handle update failed", "update_id", update.UpdateID, "err", er)

	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message:  "Update",
		Level:    sentry.LevelInfo,
		Data:     structs.Map(update),
		Category: "bot",
	})

	sentry.CaptureException(er)
}
