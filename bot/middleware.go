package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
)

func newAuthMiddleware(srv *service.Auth) tg.Middleware {
	return func(next tg.Handler) tg.Handler {
		return tg.HandlerFunc(func(ctx context.Context, update *tgbotapi.Update) error {
			var tgUser *tgbotapi.User

			switch {
			case update.Message != nil:
				tgUser = update.Message.From
			case update.EditedMessage != nil:
				tgUser = update.EditedMessage.From
			case update.CallbackQuery != nil:
				tgUser = update.CallbackQuery.From
			default:
				log.Warn(ctx, "unsupported update", "id", update.UpdateID)
				return nil
			}

			if tgUser.UserName != "" {
				ctx = log.With(ctx, "user", fmt.Sprintf("%s#%d", tgUser.UserName, tgUser.ID))
			} else {
				ctx = log.With(ctx, "user", fmt.Sprintf("#%d", tgUser.ID))
			}

			user, err := srv.Auth(ctx, &service.UserInfo{
				ID:           tgUser.ID,
				FirstName:    tgUser.FirstName,
				LastName:     tgUser.LastName,
				Username:     tgUser.UserName,
				LanguageCode: tgUser.LanguageCode,
			})

			if err != nil {
				return errors.Wrap(err, "auth service")
			}

			ctx = withUser(ctx, user)

			return next.HandleUpdate(ctx, update)
		})
	}
}
