package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lithammer/dedent"
)

var (
	textSettingsChannelsAndChats = dedent.Dedent(`
		⚙️ __*Настройки*__ / 📢 __*Каналы и чаты*__
		
		В данном разделе вы можете управлять подключенными чатами и каналами\. 
		Это эффективный инструмент для увелечения конверсии, так как позволяет ограничить скачивание файла только подписчиками вашего канала или чата\.
	`)

	textSettingsChannelsAndChatsConnect = dedent.Dedent(`
		⚙️ __*Настройки*__ / 📢 __*Каналы и чаты*__ / __*Подключить*__

		Чтобы добавить канал или чат, тебе нужно выполнить следующие действия:

		1\. Добавьте @%s в администратраторы канала или чата с минимальными правами \(например "Add User"\)\.
		2\. Отправь мне @username или приватную ссылку на канал или чат, так же ты можешь переслать любое сообщение из канала\.
	`)

	textSettingsChannelsAndChatsButtonConnect = "+ Подключить"
	callbackSettingsChannelsAndChatsConnect   = "settings:channels-and-chats:connect"
)

func (bot *Bot) newSettingsChannelsAndChatsMessageEdit(ctx context.Context, chatID int64, msgID int) tgbotapi.EditMessageTextConfig {
	answ := tgbotapi.NewEditMessageText(chatID, msgID, textSettingsChannelsAndChats)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				textSettingsChannelsAndChatsButtonConnect,
				callbackSettingsChannelsAndChatsConnect,
			),
			tgbotapi.NewInlineKeyboardButtonData(
				textCommonBack,
				callbackSettings,
			),
		),
	)

	answ.ReplyMarkup = &markup
	answ.ParseMode = "MarkdownV2"

	return answ
}

func (bot *Bot) onSettingsChannelsAndChats(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {
	go bot.answerCallbackQuery(ctx, cbq, "")
	edit := bot.newSettingsChannelsAndChatsMessageEdit(ctx, cbq.Message.Chat.ID, cbq.Message.MessageID)
	return bot.send(ctx, edit)
}

func (bot *Bot) newSettingsChannelsAndChatsConnectEdit(ctx context.Context, cid int64, mid int) tgbotapi.EditMessageTextConfig {
	text := fmt.Sprintf(textSettingsChannelsAndChatsConnect, escapeMarkdown(bot.client.Self.UserName))
	answ := tgbotapi.NewEditMessageText(cid, mid, text)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				textCommonBack,
				callbackSettingsChannelsAndChats,
			),
		),
	)

	answ.ReplyMarkup = &markup
	answ.ParseMode = "MarkdownV2"

	return answ
}

func (bot *Bot) onSettingsChannelsAndChatsConnect(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {
	go bot.answerCallbackQuery(ctx, cbq, "")
	edit := bot.newSettingsChannelsAndChatsConnectEdit(ctx, cbq.Message.Chat.ID, cbq.Message.MessageID)
	return bot.send(ctx, edit)
}
