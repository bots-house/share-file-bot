package bot

import (
	"context"
	"fmt"

	"github.com/bots-house/share-file-bot/bot/state"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
	"github.com/friendsofgo/errors"
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

	textSettingsChannelsAndChatsConnectForwardNotFromChannel = "⚠️ Пересланное сообщение не является сообщением из канала, для подключения группы или супергруппы отправь мне ее @username или приватную ссылку"
	textSettingsChannelsAndChatsConnectNotJoinLink           = "⚠️ Ссылка не является пригласительной, она должна выглядить как-то так: `t.me/joinchat/...`"
	textSettingsChannelsAndChatsConnectNotValid              = "⚠️ Для подключения канала или чата отправь мне его @username, приватную ссылку или перешли мне любое сообщение из канала"
	textSettingsChannelsAndChatsConnectIsPrivate             = "⚠️ Нужно отправить @username или приватную ссылку на канал или чат, ты скинул пользователя :)"
	textSettingsChannelsAndChatsConnectNotFound              = "⚠️ Чат не найден или бот не является админом, добавь бота в администраторы и повтори запрос"
	textSettingsChannelsAndChatsConnectBotIsNotAdmin         = "⚠️ Бот не установлен администратором чата или канала, добавьте его в администраторы с минимальными правами"
	textSettingsChannelsAndChatsConnectUserIsNotAdmin        = "⚠️ Ты не являешся администратором данного чата / канала"

	textSettingsChannelsAndChatsConnectNotValidButtonCancel = "Я передумал"
	textSettingsChannelsAndChatsButtonConnect               = "+ Подключить"
	callbackSettingsChannelsAndChatsConnect                 = "settings:channels-and-chats:connect"
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
	user := getUserCtx(ctx).ID

	go bot.answerCallbackQuery(ctx, cbq, "")

	if err := bot.state.Del(ctx, user); err != nil {
		return errors.Wrap(err, "delete state")
	}

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
	user := getUserCtx(ctx)

	if err := bot.state.Set(ctx, user.ID, state.SettingsChannelsAndChatsConnect); err != nil {
		return errors.Wrap(err, "can't set state of user")
	}

	edit := bot.newSettingsChannelsAndChatsConnectEdit(ctx, cbq.Message.Chat.ID, cbq.Message.MessageID)
	return bot.send(ctx, edit)
}

func (bot *Bot) onSettingsChannelsAndChatsConnectState(ctx context.Context, msg *tgbotapi.Message) error {

	var identity service.ChatIdentity

	switch {
	// handle forward
	case msg.ForwardDate != 0:
		if msg.ForwardFromChat == nil {
			reply := bot.newReplyMsg(msg, textSettingsChannelsAndChatsConnectForwardNotFromChannel)
			return bot.send(ctx, reply)
		}

		identity = service.NewChatIdentityFromID(msg.ForwardFromChat.ID)

	// handle @username
	case msg.Entities != nil && getFirstMentionEntity(*msg.Entities) != nil:
		entity := getFirstMentionEntity(*msg.Entities)
		username := msg.Text[entity.Offset:entity.Length]

		identity = service.NewChatIdentityFromUsername(username)

	// handle join link
	case msg.Entities != nil && getFirstLinkEntity(*msg.Entities) != nil:
		entity := getFirstLinkEntity(*msg.Entities)
		url := string([]rune(msg.Text)[entity.Offset:entity.Length])

		encodedPayload := tg.ParseJoinLink(url)

		if encodedPayload == "" {
			reply := bot.newReplyMsg(msg, textSettingsChannelsAndChatsConnectNotJoinLink)
			return bot.send(ctx, reply)
		}

		payload, err := tg.DecodeJoinLinkPayload(encodedPayload)
		if err != nil {
			return errors.Wrap(err, "decode join link payload")
		}

		identity.ID = payload.BotChatID()

	// unknown input
	default:
		answ := bot.newAnswerMsg(msg, textSettingsChannelsAndChatsConnectNotValid)
		answ.ParseMode = "MarkdownV2"
		answ.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					textSettingsChannelsAndChatsConnectNotValidButtonCancel,
					callbackSettingsChannelsAndChats,
				),
			),
		)
		return bot.send(ctx, answ)
	}

	user := getUserCtx(ctx)

	_, err := bot.chatSrv.Add(ctx, user, identity)

	switch {
	case err == service.ErrChatIsUser:
		return bot.sendText(ctx, user.ID, textSettingsChannelsAndChatsConnectIsPrivate)
	case err == service.ErrChatNotFoundOrBotIsNotAdmin:
		return bot.sendText(ctx, user.ID, textSettingsChannelsAndChatsConnectNotFound)
	case err == service.ErrBotIsNotChatAdmin:
		return bot.sendText(ctx, user.ID, textSettingsChannelsAndChatsConnectBotIsNotAdmin)
	case err == service.ErrUserIsNotChatAdmin:
		return bot.sendText(ctx, user.ID, textSettingsChannelsAndChatsConnectUserIsNotAdmin)
	case err != nil:
		return errors.Wrap(err, "add chat")
	}

	return nil
}
