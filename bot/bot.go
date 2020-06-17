package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"github.com/tomasen/realip"
)

type Bot struct {
	client  *tgbotapi.BotAPI
	authSrv *service.Auth
	handler tg.Handler
}

func New(token string, authSrv *service.Auth) (*Bot, error) {
	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "create bot api")
	}

	bot := &Bot{
		client:  client,
		authSrv: authSrv,
	}

	// bot.client.Debug = true

	bot.initHandler()

	return bot, nil
}

func (bot *Bot) SetWebhookIfNeed(ctx context.Context, domain string, path string) error {
	webhook, err := bot.client.GetWebhookInfo()
	if err != nil {
		return errors.Wrap(err, "get webhook info")
	}
	endpoint := strings.Join([]string{domain, path}, "/")

	if webhook.URL != endpoint {
		u, err := url.Parse(endpoint)
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

func (bot *Bot) onUpdate(ctx context.Context, update *tgbotapi.Update) error {
	// spew.Dump(update)

	if msg := update.Message; msg != nil {
		// handle command
		switch msg.Command() {
		case "start":
			return bot.onStart(ctx, msg)
		case "help":
			return bot.onHelp(ctx, msg)
		}

		// handle other
		if msg.Document == nil {
			return bot.onNotDocument(ctx, msg)
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
		log.Error(ctx, "handle update failed", "update_id", update.UpdateID, "err", err)
		http.Error(w, fmt.Sprintf("handle error: %s", err), http.StatusInternalServerError)
		return
	}
}
