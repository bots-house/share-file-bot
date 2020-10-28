package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/friendsofgo/errors"
	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
)

const refDeepLinkPrefix = "ref_"

func getRefFromMessage(msg *tgbotapi.Message) string {
	if msg != nil && msg.Command() == "start" {
		args := msg.CommandArguments()

		if strings.HasPrefix(args, refDeepLinkPrefix) {
			return strings.TrimPrefix(args, refDeepLinkPrefix)
		}
	}

	return ""
}

func newAuthMiddleware(srv *service.Auth) tg.Middleware {
	return func(next tg.Handler) tg.Handler {
		return tg.HandlerFunc(func(ctx context.Context, update *tgbotapi.Update) error {
			var (
				tgUser *tgbotapi.User
			)

			switch {
			case update.Message != nil:
				tgUser = update.Message.From
			case update.EditedMessage != nil:
				tgUser = update.EditedMessage.From
			case update.CallbackQuery != nil:
				tgUser = update.CallbackQuery.From
			case update.ChannelPost != nil && update.ChannelPost.NewChatTitle != "":
				tgUser = nil
			default:
				log.Warn(ctx, "unsupported update", "id", update.UpdateID)
				return nil
			}

			if tgUser != nil {
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
					Ref:          getRefFromMessage(update.Message),
				})

				if err != nil {
					return errors.Wrap(err, "auth service")
				}

				sentry.AddBreadcrumb(&sentry.Breadcrumb{
					Message: "Authenticated",
					Level:   sentry.LevelInfo,
					Data:    structs.Map(user),
				})

				ctx = withUser(ctx, user)

				sentry.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetUser(sentry.User{
						ID:       strconv.Itoa(int(user.ID)),
						Username: user.Username.String,
					})
				})
			}

			return next.HandleUpdate(ctx, update)
		})
	}
}
