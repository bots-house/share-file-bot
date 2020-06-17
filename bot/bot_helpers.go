package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (bot *Bot) send(ctx context.Context, msg *tgbotapi.MessageConfig) error {
	// spew.Dump(msg)
	_, err := bot.client.Send(msg)
	return err
}

func (bot *Bot) newAnswerMsg(ctx context.Context, msg *tgbotapi.Message, text string) *tgbotapi.MessageConfig {
	result := tgbotapi.NewMessage(
		int64(msg.From.ID),
		text,
	)

	result.ParseMode = tgbotapi.ModeMarkdown

	return &result
}
