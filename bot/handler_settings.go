package bot

import (
	"context"

	"github.com/friendsofgo/errors"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lithammer/dedent"
)

var (
	textSettings = dedent.Dedent(`
        ⚙️ __*Настройки*__

		• _Длинные ID_ — бот будет генерировать максимально возможные по длине ссылки, идеально для личных файлов\. Длинные ссылки буду генерироватся только для новых документов\.

		• _Каналы и чаты_ — управление каналами и чата подключенными к боту в качестве ограничителя при скачивании ваших файлов\.
    `)

	textCommonBack       = "« Назад"
	textCommonDisconnect = "Отключить"
	textCommonYesIamSure = "Да, я уверен"

	textSettingsButtonLongIDs              = "Длинные ID"
	textSettingsButtonLongIDsEnabledAlert  = "Генериация длинных ссылок включена"
	textSettingsButtonLongIDsDisabledAlert = "Генериация длинных ссылок выключена"

	textSettingsButtonChannelsAndChats = "📢 Каналы и чаты"

	callbackSettings        = "settings"
	callbackSettingsLongIDs = "settings:toggle-long-ids"
)

func addIsEnabledEmoji(v bool, text string) string {
	if v {
		return "✅ " + text
	}

	return text
}

func (bot *Bot) newSettingsMenuMessageReplyMarkup(longIDs bool) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				addIsEnabledEmoji(longIDs, textSettingsButtonLongIDs),
				callbackSettingsLongIDs,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				textSettingsButtonChannelsAndChats,
				callbackSettingsChannelsAndChats,
			),
		),
	)
}

func (bot *Bot) newSettingsMenuMessage(msg *tgbotapi.Message, user *core.User) *tgbotapi.MessageConfig {
	answ := bot.newAnswerMsg(msg, textSettings)
	answ.ReplyMarkup = bot.newSettingsMenuMessageReplyMarkup(user.Settings.LongIDs)
	answ.ParseMode = "MarkdownV2"

	return answ
}

func (bot *Bot) newSettingsMenuMessageEdit(msg *tgbotapi.Message, user *core.User) tgbotapi.EditMessageTextConfig {
	answ := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, textSettings)

	markup := bot.newSettingsMenuMessageReplyMarkup(user.Settings.LongIDs)

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

func (bot *Bot) onSettingsCallbackQuery(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {

	user := getUserCtx(ctx)
	answ := bot.newSettingsMenuMessageEdit(cbq.Message, user)
	return bot.send(ctx, answ)
}
