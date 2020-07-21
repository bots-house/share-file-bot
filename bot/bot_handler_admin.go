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

	text := strings.Join([]string{
		"*#cтатистика*",
		"",
		fmt.Sprintf("*Пользователи*: `%d`", stats.Users),
		fmt.Sprintf("*Документы*: `%d`", stats.Documents),
		fmt.Sprintf("*Загрузки*: `%d`", stats.Downloads),
	}, "\n")

	return bot.send(ctx, bot.newAnswerMsg(msg, text))
}
