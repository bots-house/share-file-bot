package bot

import (
	"context"
	"encoding/json"
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

func extractRefFromMsg(msg *tgbotapi.Message) string {
	if msg != nil && msg.Command() == "start" {
		args := msg.CommandArguments()

		if !strings.HasPrefix(args, refDeepLinkPrefix) {
			return ""
		}

		ref := strings.TrimPrefix(args, refDeepLinkPrefix)

		if !strings.Contains(ref, "-") {
			msg.Text = "/start"
			return ref
		}

		items := strings.Split(ref, "-")

		if len(items) > 1 {
			msg.Text = "/start " + items[1]
		}

		return items[0]
	}

	return ""
}

func serializeStruct(v interface{}) map[string]interface{} {
	body, err := json.Marshal(v)
	if err != nil {
		return structs.Map(v)
	}

	result := map[string]interface{}{}

	if err := json.Unmarshal(body, &result); err != nil {
		return structs.Map(v)
	}

	return result
}

func newAuthMiddleware(srv *service.Auth) tg.Middleware {
	return func(next tg.Handler) tg.Handler {
		return tg.HandlerFunc(func(ctx context.Context, update *tgbotapi.Update) error {
			withSentryHub(ctx, func(hub *sentry.Hub) {
				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Message:  "Update",
					Level:    sentry.LevelInfo,
					Data:     serializeStruct(update),
					Category: "middleware",
				}, nil)
			})

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

				ref := extractRefFromMsg(update.Message)

				user, err := srv.Auth(ctx, &service.UserInfo{
					ID:           tgUser.ID,
					FirstName:    tgUser.FirstName,
					LastName:     tgUser.LastName,
					Username:     tgUser.UserName,
					LanguageCode: tgUser.LanguageCode,
					Ref:          ref,
				})

				if err != nil {
					return errors.Wrap(err, "auth service")
				}

				withSentryHub(ctx, func(hub *sentry.Hub) {
					ctx = withUser(ctx, user)

					hub.AddBreadcrumb(&sentry.Breadcrumb{
						Message:  "User",
						Level:    sentry.LevelInfo,
						Data:     serializeStruct(user),
						Category: "auth",
					}, nil)

					hub.ConfigureScope(func(scope *sentry.Scope) {
						scope.SetUser(sentry.User{
							ID:       strconv.Itoa(int(user.ID)),
							Username: user.Username.String,
						})
					})
				})
			}

			return next.HandleUpdate(ctx, update)
		})
	}
}

func withSentryHub(ctx context.Context, do func(hub *sentry.Hub)) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		do(hub)
	}
}
