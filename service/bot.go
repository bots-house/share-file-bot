package service

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
)

// WebhookBuilder define function signature for build URL of bot webhook
type WebhookBuilder func(token string) string

// BotService it's manager of user bot.
type BotService struct {
	Bot   core.BotStore
	Txier store.Txier

	WebhookBuilder WebhookBuilder
}

// GetBots returns bots of user.
func (srv *BotService) GetBots(
	ctx context.Context,
	user *core.User,
) ([]*core.Bot, error) {
	bots, err := srv.Bot.Query().OwnerID(user.ID).All(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "query user bots")
	}

	return bots, nil
}

// FullBot define details representation of bot.
type FullBot struct {
	*core.Bot
}

var (
	// ErrBotAlreadyConnected returned by BotService.Connect method when bot is already connect.
	ErrBotAlreadyConnected = errors.New("bot already connected")
)

// BotAlreadyUsedError returned by BotService.Connect method when bot is already used by another service.
type BotAlreadyUsedError struct {
	// Host contains webhook host of domain
	Host string
}

func (err *BotAlreadyUsedError) Error() string {
	return "bot alredy used by " + err.Host
}

// Connect bot. Do following:
//  - check if token is valid;
//  - check if bot is not already connected;
//  - check if bot is alredy used by another service and ignore if forceReuse = true;
//  - save bot to store;
func (srv *BotService) Connect(
	ctx context.Context,
	user *core.User,
	token string,
	forceReuse bool,
) (*FullBot, error) {
	// create new api instance (internally call getMe)
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "get me")
	}

	me := api.Self

	// check if bot is not already connected
	total, err := srv.Bot.Query().ID(core.BotID(me.ID)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "count bots with same id")
	}

	if total > 0 {
		return nil, ErrBotAlreadyConnected
	}

	// check if bot is not used by another service
	webhook, err := api.GetWebhookInfo()
	if err != nil {
		return nil, errors.Wrap(err, "can't get webhook info")
	}

	if wh := webhook.URL; wh != "" && !forceReuse {
		wh = normalizeWebhookURL(wh)

		uri, err := url.ParseRequestURI(wh)
		if err != nil {
			return nil, errors.Wrap(err, "parse webhook uri")
		}

		return nil, &BotAlreadyUsedError{Host: uri.Host}
	}

	// set webhook of bot
	newWebhook := tgbotapi.NewWebhook(srv.WebhookBuilder(token))
	newWebhook.MaxConnections = 100

	bot := &core.Bot{
		ID:       core.BotID(me.ID),
		Username: me.UserName,
		Token:    token,
		OwnerID:  user.ID,
		LinkedAt: time.Now(),
	}

	if err := srv.Txier(ctx, func(ctx context.Context) error {
		if err := srv.Bot.Add(ctx, bot); err != nil {
			return errors.Wrap(err, "add to store")
		}

		_, err = api.SetWebhook(newWebhook)
		if err != nil {
			return errors.Wrap(err, "set webhook")
		}

		return err
	}); err != nil {
		return nil, err
	}

	return &FullBot{Bot: bot}, nil
}

// normalizeWebhookURL return u with protocol for correct parsing.
// You can set webhook without URL and it's break parsing of host.
func normalizeWebhookURL(u string) string {
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return u
	}

	return "http://" + u
}
