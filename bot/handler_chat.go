package bot

import (
	"context"

	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (bot *Bot) onChatNewTitle(ctx context.Context, msg *tgbotapi.Message) error {
	if err := bot.chatSrv.UpdateTitle(ctx, msg.Chat.ID, msg.NewChatTitle); err != nil {
		return errors.Wrap(err, "update title")
	}

	return nil
}
