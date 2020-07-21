package bot

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lithammer/dedent"
)

var (
	textSettings = dedent.Dedent(`
        ⚙️ __*Настройки*__

        _Длинные ID_ — бот будет генерировать максимально возможные по длине ссылки, идеально для личных файлов\. Длинные ссылки буду генерироватся только для новых документов\.
    `)

	textSettingsButtonLongIDs              = "Длинные ID"
	textSettingsButtonLongIDsEnabledAlert  = "Генериация длинных ссылок включена"
	textSettingsButtonLongIDsDisabledAlert = "Генериация длинных ссылок выключена"

	callbackSettingsLongIDs = "settings:toggle-long-ids"
)

func addIsEnabledEmoji(v bool, text string) string {
	if v {
		return "✅ " + text
	}

	return text
}

func (bot *Bot) newSettingsMenuMessage(msg *tgbotapi.Message, user *core.User) *tgbotapi.MessageConfig {
	answ := bot.newAnswerMsg(msg, textSettings)
	answ.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				addIsEnabledEmoji(user.Settings.LongIDs, textSettingsButtonLongIDs),
				callbackSettingsLongIDs,
			),
		),
	)
	answ.ParseMode = "MarkdownV2"

	return answ
}

func (bot *Bot) newSettingsMenuMessageEdit(msg *tgbotapi.Message, user *core.User) tgbotapi.EditMessageTextConfig {
	answ := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, textSettings)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				addIsEnabledEmoji(user.Settings.LongIDs, textSettingsButtonLongIDs),
				callbackSettingsLongIDs,
			),
		),
	)

	answ.ReplyMarkup = &markup
	answ.ParseMode = "MarkdownV2"

	return answ
}

func (bot *Bot) onSettingsToggleLongIDsCBQ(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {
	user := getUserCtx(ctx)

	isEnabled, err := bot.authSrv.SettingsToggleLongIDs(ctx, user)
	if err != nil {
		return errors.Wrap(err, "toggle settings long ids")
	}

	var answer string

	if isEnabled {
		answer = textSettingsButtonLongIDsEnabledAlert
	} else {
		answer = textSettingsButtonLongIDsDisabledAlert
	}

	go func() {
		if err := bot.answerCallbackQuery(ctx, cbq, answer); err != nil {
			log.Warn(ctx, "cant answer inline query in onSettingsToggleLongIDsCBQ", "err", err)
		}
	}()

	answ := bot.newSettingsMenuMessageEdit(cbq.Message, user)
	return bot.send(ctx, answ)
}

func (bot *Bot) onSettings(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)

	answ := bot.newSettingsMenuMessage(msg, user)

	return bot.send(ctx, answ)
}
