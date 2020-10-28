package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/service"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (bot *Bot) onAdmin(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)

	stats, err := bot.adminSrv.SummaryStats(ctx, user)
	if errors.Cause(err) == service.ErrUserIsNotAdmin {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "summary stats")
	}

	lines := []string{
		"*__Общая__*",
		"",
		fmt.Sprintf("*Пользователи*: `%d`", stats.Users),
		fmt.Sprintf("*Файлы*: `%d`", stats.Files),
		fmt.Sprintf("*Загрузки*: `%d`", stats.Downloads),
		fmt.Sprintf("*Чаты*: `%d`", stats.Chats),
		"",
		"*__Источники__*",
		"",
	}

	for _, item := range stats.UsersByRefs {
		var key string

		if item.Ref.Valid {
			key = item.Ref.String
		} else {
			key = "null"
		}

		lines = append(lines,
			fmt.Sprintf("*%s*: `%d`", key, item.Count),
		)
	}

	text := strings.Join(lines, "\n")

	out := tgbotapi.NewMessage(msg.Chat.ID, text)
	out.ParseMode = mdv2

	return bot.send(ctx, out)
}
