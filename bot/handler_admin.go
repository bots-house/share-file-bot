package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/service"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (bot *Bot) onAdminStats(ctx context.Context, msg *tgbotapi.Message) error {
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

func (bot *Bot) onAdminStatsRef(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)

	ref := msg.CommandArguments()
	if ref == "" {
		return service.ErrArgsAreEmpty
	}

	summary, err := bot.adminSrv.SummaryRefStats(ctx, user, ref)
	if err != nil {
		return errors.Wrap(err, "get summary ref stats")
	}

	lines := []string{
		fmt.Sprintf("*__%s__*", ref),
		"",
		fmt.Sprintf("*Регистрации*: `%d`", summary.ClickedOnStart),
		fmt.Sprintf("*Подключили чат*: `%d`", summary.ConnectedChat),
		fmt.Sprintf("*Загрузили файл*: `%d`", summary.DownloadCount),
		fmt.Sprintf("*Установили ограничение на загрузку*: `%d`", summary.SetRestriction),
		fmt.Sprintf("*Получили загрузок*: `%d`", summary.UploadedFile),
	}

	text := strings.Join(lines, "\n")

	out := tgbotapi.NewMessage(msg.Chat.ID, text)
	out.ParseMode = mdv2

	return bot.send(ctx, out)
}
