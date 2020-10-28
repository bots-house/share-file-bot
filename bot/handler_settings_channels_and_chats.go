package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/bot/state"
	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lithammer/dedent"
)

func join(vs ...string) string {
	return strings.Join(vs, "\n")
}

var (
	textSettingsChannelsAndChats = dedent.Dedent(`
		⚙️ __*Настройки*__ / 📢 __*Каналы и чаты*__
		
		В данном разделе вы можете управлять подключенными чатами и каналами\. 
		Это эффективный инструмент для увелечения конверсии, так как позволяет ограничить скачивание файла только подписчиками вашего канала или чата\.
	`)

	textSettingsChannelsAndChatsConnect = dedent.Dedent(`
		⚙️ __*Настройки*__ / 📢 __*Каналы и чаты*__ / __*Подключить*__

		Чтобы добавить канал или чат, тебе нужно выполнить следующие действия:

		1\. Добавьте @%s в администратраторы канала или чата с правами «Добавление подписчиков» \(Add User\)\.
		2\. Отправь мне @username или приватную ссылку на канал или чат, так же ты можешь переслать любое сообщение из канала\.
	`)

	textSettingsChannelsAndChatsDetails = join(
		"⚙️ __*Настройки*__ / 📢 __*Каналы и чаты*__ / __*%s*__",
		"",
		"*ID:* `%d`",
		"*Тип:* `%s`",
		"",
		"📈 __Статистика__",
		"",
		"*Файлов:* `%d`",
		"*Загрузок с подпиской:* `%d`",
		"*Загрузок с новой подпиской*: `%d`",
	)

	textSettingsChannelsAndChatsDelete = dedent.Dedent(`
		⚙️ __*Настройки*__ / 📢 __*Каналы и чаты*__ / __*%s*__

		Уверены что хотите отвзять этот канал/группу? 
		Файл для скачивания которых требовалась подписка на этот канал/группу, станут доступны без нее\.
		Это действие нельзя будет отменить\. 
	`)

	textSettingsChannelsAndChatsConnectForwardNotFromChannel = "⚠️ Пересланное сообщение не является сообщением из канала, для подключения группы или супергруппы отправь мне ее @username или приватную ссылку"
	textSettingsChannelsAndChatsConnectNotJoinLink           = "⚠️ Ссылка не является пригласительной, она должна выглядить как-то так: `t.me/joinchat/...`"
	textSettingsChannelsAndChatsConnectNotValid              = "⚠️ Для подключения канала или чата отправь мне его @username, приватную ссылку или перешли мне любое сообщение из канала"
	textSettingsChannelsAndChatsConnectIsPrivate             = "⚠️ Нужно отправить @username или приватную ссылку на канал или чат, ты скинул пользователя :)"
	textSettingsChannelsAndChatsConnectNotFound              = "⚠️ Чат не найден или бот не является админом, добавь бота в администраторы и повтори запрос"
	textSettingsChannelsAndChatsConnectBotIsNotAdmin         = "⚠️ Бот не установлен администратором чата или канала, добавьте его в администраторы с правами «Добавление подписчиков» (Add User)"
	textSettingsChannelsAndChatsConnectUserIsNotAdmin        = "⚠️ Ты не являешся администратором данного чата / канала"
	textSettingsChannelsAndChatsConnectBotIsNotEnoughRights  = "⚠️ Бот установлен администратором чата / канала, но ему не хватает прав «Добавление подписчиков» (Add User)"

	textSettingsChannelsAndChatsConnectNotValidButtonCancel = "Я передумал"
	textSettingsChannelsAndChatsButtonConnect               = "+ Подключить"

	callbackSettingsChannelsAndChats              = "settings:channels-and-chats"
	callbackSettingsChannelsAndChatsConnect       = "settings:channels-and-chats:connect"
	callbackSettingsChannelsAndChatsDetails       = "settings:channels-and-chats:%d"
	callbackSettingsChannelsAndChatsDelete        = "settings:channels-and-chats:%d:delete"
	callbackSettingsChannelsAndChatsDeleteConfirm = "settings:channels-and-chats:%d:delete:confirm"
)

func (bot *Bot) newSettingsChannelsAndChatsMessageEdit(
	chatID int64,
	msgID int,
	chats []*core.Chat,
) tgbotapi.EditMessageTextConfig {
	answ := tgbotapi.NewEditMessageText(chatID, msgID, textSettingsChannelsAndChats)

	chatRows := make([][]tgbotapi.InlineKeyboardButton, len(chats))

	for i, chat := range chats {
		cbData := fmt.Sprintf(callbackSettingsChannelsAndChatsDetails, chat.ID)
		chatRows[i] = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(chat.Title, cbData),
		)
	}

	chatRows = append(chatRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			textCommonBack,
			callbackSettings,
		),
		tgbotapi.NewInlineKeyboardButtonData(
			textSettingsChannelsAndChatsButtonConnect,
			callbackSettingsChannelsAndChatsConnect,
		),
	))

	markup := tgbotapi.NewInlineKeyboardMarkup(
		chatRows...,
	)

	answ.ReplyMarkup = &markup
	answ.ParseMode = mdv2

	return answ
}

func (bot *Bot) onSettingsChannelsAndChats(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {
	user := getUserCtx(ctx)

	go func() {
		_ = bot.answerCallbackQuery(ctx, cbq, "")
	}()

	if err := bot.state.Del(ctx, user.ID); err != nil {
		return errors.Wrap(err, "delete state")
	}

	chats, err := bot.chatSrv.GetChats(ctx, user)
	if err != nil {
		return errors.Wrap(err, "get chats")
	}

	edit := bot.newSettingsChannelsAndChatsMessageEdit(cbq.Message.Chat.ID, cbq.Message.MessageID, chats)
	return bot.send(ctx, edit)
}

func (bot *Bot) newSettingsChannelsAndChatsConnectEdit(cid int64, mid int) tgbotapi.EditMessageTextConfig {
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
	answ.ParseMode = mdv2

	return answ
}

func (bot *Bot) onSettingsChannelsAndChatsConnect(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {
	go func() {
		_ = bot.answerCallbackQuery(ctx, cbq, "")
	}()

	user := getUserCtx(ctx)

	if err := bot.state.Set(ctx, user.ID, state.SettingsChannelsAndChatsConnect); err != nil {
		return errors.Wrap(err, "can't set state of user")
	}

	edit := bot.newSettingsChannelsAndChatsConnectEdit(cbq.Message.Chat.ID, cbq.Message.MessageID)
	return bot.send(ctx, edit)
}

func getChatTypeRussian(typ core.ChatType) string {
	switch typ {
	case core.ChatTypeChannel:
		return "канал"
	case core.ChatTypeSuperGroup:
		return "супергруппа"
	case core.ChatTypeGroup:
		return "группа"
	default:
		return "неизвестно"
	}
}

func (bot *Bot) newSettingsChannelsAndChatsDetailsEdit(
	cid int64, mid int,
	chat *service.FullChat,
) *tgbotapi.EditMessageTextConfig {
	stats := chat.GetStats()

	text := fmt.Sprintf(
		textSettingsChannelsAndChatsDetails,
		escapeMarkdown(chat.Title),
		chat.TelegramID,
		getChatTypeRussian(chat.Type),
		chat.Files,
		stats.WithSubscription,
		stats.NewSubscription,
	)

	answ := tgbotapi.NewEditMessageText(cid, mid, text)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				textCommonBack,
				callbackSettingsChannelsAndChats,
			),
			tgbotapi.NewInlineKeyboardButtonData(
				textCommonDisconnect,
				fmt.Sprintf(callbackSettingsChannelsAndChatsDelete, chat.ID),
			),
		),
	)

	answ.ReplyMarkup = &markup
	answ.ParseMode = mdv2

	return &answ
}

func (bot *Bot) onSettingsChannelsAndChatsDeleteConfirm(
	ctx context.Context,
	user *core.User,
	cbq *tgbotapi.CallbackQuery,
	id core.ChatID,
) error {
	chat, err := bot.chatSrv.GetChat(ctx, user, id)
	if err != nil {
		return errors.Wrap(err, "Chat.GetChat")
	}

	if err := bot.chatSrv.DisconnectChat(ctx, user, chat.ID, false); err != nil {
		return errors.Wrap(err, "service disconnect chat")
	}

	go func() {
		_ = bot.answerCallbackQuery(ctx, cbq, "Канал/группа отключена")
	}()

	chats, err := bot.chatSrv.GetChats(ctx, user)
	if err != nil {
		return errors.Wrap(err, "service get chats")
	}

	edit := bot.newSettingsChannelsAndChatsMessageEdit(cbq.Message.Chat.ID, cbq.Message.MessageID, chats)

	return bot.send(ctx, edit)
}

func (bot *Bot) onSettingsChannelsAndChatsDelete(
	ctx context.Context,
	user *core.User,
	cbq *tgbotapi.CallbackQuery,
	id core.ChatID,
) error {
	chat, err := bot.chatSrv.GetChat(ctx, user, id)
	if err != nil {
		return errors.Wrap(err, "Chat.GetChat")
	}

	return bot.send(ctx, bot.newSettingsChannelsAndChatsDeleteEdit(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		chat,
	))
}

func (bot *Bot) newSettingsChannelsAndChatsDeleteEdit(
	cid int64,
	mid int,
	chat *service.FullChat,
) tgbotapi.EditMessageTextConfig {
	text := fmt.Sprintf(textSettingsChannelsAndChatsDelete, escapeMarkdown(chat.Title))
	answ := tgbotapi.NewEditMessageText(
		cid,
		mid,
		text,
	)

	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				textCommonBack,
				fmt.Sprintf(callbackSettingsChannelsAndChatsDetails, chat.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				textCommonYesIamSure,
				fmt.Sprintf(callbackSettingsChannelsAndChatsDeleteConfirm, chat.ID),
			),
		),
	)

	answ.ParseMode = mdv2

	answ.ReplyMarkup = &replyMarkup

	return answ
}

func (bot *Bot) onSettingsChannelsAndChatsDetails(
	ctx context.Context,
	user *core.User,
	cbq *tgbotapi.CallbackQuery,
	id core.ChatID,
) error {
	chat, err := bot.chatSrv.GetChat(ctx, user, id)
	if err != nil {
		return errors.Wrap(err, "Chat.GetChat")
	}

	edit := bot.newSettingsChannelsAndChatsDetailsEdit(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		chat,
	)

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
		answ.ParseMode = mdv2
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
	case err == service.ErrBotNotEnoughRights:
		return bot.sendText(ctx, user.ID, textSettingsChannelsAndChatsConnectBotIsNotEnoughRights)
	case err == service.ErrUserIsNotChatAdmin:
		return bot.sendText(ctx, user.ID, textSettingsChannelsAndChatsConnectUserIsNotAdmin)
	case err != nil:
		return errors.Wrap(err, "add chat")
	}

	out := tgbotapi.NewMessage(
		msg.Chat.ID,
		fmt.Sprintln("Канал / супергруппа подключена!"),
	)

	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(textCommonBack, callbackSettingsChannelsAndChats),
		),
	)

	return bot.send(ctx, out)
}
