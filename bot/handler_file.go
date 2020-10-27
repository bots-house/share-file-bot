package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/service"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lithammer/dedent"
)

const tgDomain = "t.me"

const (
	callbackFileRestrictions          = "file:%d:restrictions"
	callbackFileRestrictionsChat      = "file:%d:restrictions:chat-subscription:%d:toggl"
	callbackFileRestrictionsChatCheck = "file:%d:restrictions:chat:check"
)

var (
	textFileRestrictions = dedent.Dedent(`
		С помощью данного инструмента вы можете ограничить доступ к файлу только подписчикам вашего канала / супергруппы\.
		Перед каждым скачиванием бот будет проверять наличие подписки и только после этого выдавать доступ к файлу\. 
		
		_Для подключения каналов перейдите в настройки \(/settings\)\._
	`)

	textFileSubRequest = dedent.Dedent(`
		Владелец файла установил ограничение на доступ только с подпиской\. 
		Подпишись на %s и нажми кнопку *«Я подписался»*
	`)
)

func (bot *Bot) renderNotOwnedFile(msg *tgbotapi.Message, file *core.File) tgbotapi.Chattable {
	return bot.renderGenericFile(
		msg.Chat.ID,
		file.Kind,
		file.TelegramID,
		escapeMarkdown(file.Caption.String),
		tgbotapi.ModeMarkdown,
		nil,
	)
}

func (bot *Bot) renderOwnedFileCaption(file *service.OwnedFile) string {
	rows := []string{}

	if file.Caption.String != "" {
		rows = append(rows,
			fmt.Sprintf("*Описание*: %s", escapeMarkdown(file.Caption.String)),
			"",
		)
	}

	rows = append(rows,
		fmt.Sprintf("*Кол-во загрузок*: `%d`", file.Stats.Total),
		fmt.Sprintf("*Кол-во уникальных загрузок*: `%d`", file.Stats.Unique),
		"",
	)

	rows = append(rows, fmt.Sprintf("*Публичная ссылка*: https://%s/%s?start=%s",
		tgDomain,
		escapeMarkdown(bot.client.Self.UserName),
		escapeMarkdown(file.PublicID),
	))

	return strings.Join(rows, "\n")
}

func (bot *Bot) renderSubRequest(msg *tgbotapi.Message, sub *service.ChatSubRequest) tgbotapi.MessageConfig {
	var link string

	if sub.Username != "" {
		link = escapeMarkdown("@" + sub.Username)
	} else {
		link = fmt.Sprintf("[%s](%s)", sub.Title, sub.JoinLink)
	}

	text := fmt.Sprintf(textFileSubRequest, link)

	out := tgbotapi.NewMessage(msg.Chat.ID, text)
	out.ParseMode = "MarkdownV2"
	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Подписаться", sub.Link()),
			tgbotapi.NewInlineKeyboardButtonData("Я подписался", fmt.Sprintf(callbackFileRestrictionsChatCheck, sub.FileID)),
		),
	)

	return out
}

func (bot *Bot) renderOwnedFileReplyMarkup(file *service.OwnedFile) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Обновить",
				fmt.Sprintf("file:%d:refresh", file.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Удалить",
				fmt.Sprintf("file:%d:delete", file.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				addHasLockEmoji(file.Restriction.Any(), "Ограничения"),
				fmt.Sprintf(callbackFileRestrictions, file.ID),
			),
		),
	)
}

func (bot *Bot) renderGenericFile(
	chatID int64,
	fileKind core.Kind,
	fileID string,
	caption string,
	parseMode string,
	replyMarkup interface{},
) tgbotapi.Chattable {
	switch fileKind {
	case core.KindDocument:
		share := tgbotapi.NewDocumentShare(chatID, fileID)
		share.Caption = caption
		share.ReplyMarkup = replyMarkup
		share.ParseMode = parseMode
		return share
	case core.KindAnimation:
		share := tgbotapi.NewAnimationShare(chatID, fileID)
		share.Caption = caption
		share.ReplyMarkup = replyMarkup
		share.ParseMode = parseMode
		return share
	case core.KindAudio:
		share := tgbotapi.NewAudioShare(chatID, fileID)
		share.Caption = caption
		share.ReplyMarkup = replyMarkup
		share.ParseMode = parseMode
		return share
	case core.KindPhoto:
		share := tgbotapi.NewPhotoShare(chatID, fileID)
		share.Caption = caption
		share.ReplyMarkup = replyMarkup
		share.ParseMode = parseMode
		return share
	case core.KindVideo:
		share := tgbotapi.NewVideoShare(chatID, fileID)
		share.Caption = caption
		share.ReplyMarkup = replyMarkup
		share.ParseMode = parseMode
		return share
	case core.KindVoice:
		share := tgbotapi.NewVoiceShare(chatID, fileID)
		share.Caption = caption
		share.ReplyMarkup = replyMarkup
		share.ParseMode = parseMode
		return share
	default:
		return nil
	}
}

func (bot *Bot) renderOwnedFile(msg *tgbotapi.Message, file *service.OwnedFile) tgbotapi.Chattable {
	return bot.renderGenericFile(
		msg.Chat.ID,
		file.Kind,
		file.TelegramID,
		bot.renderOwnedFileCaption(file),
		tgbotapi.ModeMarkdown,
		bot.renderOwnedFileReplyMarkup(file),
	)
}

func (bot *Bot) deleteMessage(ctx context.Context, msg *tgbotapi.Message) error {
	if err := bot.send(ctx, tgbotapi.NewDeleteMessage(
		msg.Chat.ID,
		msg.MessageID,
	)); err != nil {
		log.Warn(ctx, "can't delete incoming message", "chat_id", msg.Chat.ID, "msg_id", msg.MessageID)
		return err
	}

	return nil
}

func (bot *Bot) onFile(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)

	inputFile := bot.extractInputFileFromMessage(msg)

	if inputFile == nil {
		_ = bot.sendText(ctx,
			user.ID,
			"⚠️ Упс, я не могу добавить этот файл, так как не поддерживаю его",
		)

		return core.ErrInvalidKind
	}

	// delete user message for avoid trash in history
	go func() {
		_ = bot.deleteMessage(ctx, msg)
	}()

	file, err := bot.fileSrv.AddFile(ctx, user, inputFile)

	if err != nil {
		_ = bot.sendText(ctx,
			user.ID,
			"⚠️ Что-то пошло не так при добавлении файла",
		)

		return errors.Wrap(err, "service add file")
	}

	result := bot.renderOwnedFile(msg, file)

	return bot.send(ctx, result)
}

func (bot *Bot) getFileForOwner(ctx context.Context, cbq *tgbotapi.CallbackQuery, id int) (*service.OwnedFile, error) {
	user := getUserCtx(ctx)

	doc, err := bot.fileSrv.GetFileByID(ctx, user, core.FileID(id))
	if err != nil {
		return nil, errors.Wrap(err, "get file by id")
	}

	// user is not owner of file but try to access
	if doc.OwnedFile == nil {
		if cbq != nil {
			_ = bot.answerCallbackQuery(ctx, cbq, "bad body, what you do?")
		}
		return nil, errors.New("can't manage file")
	}

	return doc.OwnedFile, nil
}

func (bot *Bot) onFileRefreshCBQ(ctx context.Context, cbq *tgbotapi.CallbackQuery, id int) error {
	doc, err := bot.getFileForOwner(ctx, cbq, id)
	if err != nil {
		return errors.Wrap(err, "get file for owner")
	}

	caption := bot.renderOwnedFileCaption(doc)

	edit := tgbotapi.NewEditMessageCaption(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		caption,
	)

	edit.ParseMode = tgbotapi.ModeMarkdown
	replyMarkup := bot.renderOwnedFileReplyMarkup(doc)
	edit.ReplyMarkup = &replyMarkup

	if err := bot.send(ctx, edit); err != nil {
		if err, ok := err.(tgbotapi.Error); ok {
			if strings.Contains(err.Message, "message is not modified:") {
				return bot.answerCallbackQuery(ctx, cbq, "🤷 Ничего не изменилось")
			}
		}
		return errors.Wrap(err, "edit message error")
	}

	return bot.answerCallbackQuery(ctx, cbq, "")
}

func (bot *Bot) onFileDeleteCBQ(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	id int,
) error {
	file, err := bot.getFileForOwner(ctx, cbq, id)
	if err != nil {
		return errors.Wrap(err, "get file for owner")
	}

	go func() {
		_ = bot.answerCallbackQuery(ctx, cbq, "")
	}()

	edit := tgbotapi.NewEditMessageCaption(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		strings.Join([]string{
			"Уверены что хотите *удалить* этот файл?",
			"",
			"Пользователи больше не смогут получить доступ к документу перейдя по ссылке.",
			"Но у пользователей уже скачавших документ, он сохранится в истории диалога с ботом.",
		}, "\n"),
	)
	edit.ParseMode = tgbotapi.ModeMarkdown

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Да, уверен",
				fmt.Sprintf("file:%d:delete:confirm", file.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Нет",
				fmt.Sprintf("file:%d:refresh", file.ID),
			),
		),
	)

	edit.ReplyMarkup = &markup

	return bot.send(ctx, edit)
}

func (bot *Bot) onFileDeleteConfirmCBQ(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	id int,
) error {
	user := getUserCtx(ctx)
	docID := core.FileID(id)

	if err := bot.fileSrv.DeleteFile(ctx, user, docID); err == core.ErrFileNotFound {
		return bot.answerCallbackQuery(ctx, cbq, "Файл не найден")
	} else if err != nil {
		return errors.Wrap(err, "service delete file")
	}

	go func() {
		_ = bot.deleteMessage(ctx, cbq.Message)
	}()

	return bot.answerCallbackQuery(ctx, cbq, "✅ Документ удален")
}

func (bot *Bot) onFileRestrictionsCBQ(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	fid int,
) error {
	go func() {
		_ = bot.answerCallbackQuery(ctx, cbq, "")
	}()

	file, err := bot.getFileForOwner(ctx, cbq, fid)
	if err != nil {
		return errors.Wrap(err, "get file for owner")
	}

	user := getUserCtx(ctx)

	chats, err := bot.chatSrv.GetChats(ctx, user)
	if err != nil {
		return errors.Wrap(err, "service get chats")
	}

	markup := bot.newFileRestrictionsReplyMarkup(file.File, chats)

	edit := tgbotapi.EditMessageCaptionConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      cbq.Message.Chat.ID,
			MessageID:   cbq.Message.MessageID,
			ReplyMarkup: markup,
		},
		ParseMode: "MarkdownV2",
		Caption:   textFileRestrictions,
	}

	return bot.send(ctx, edit)
}

func (bot *Bot) newFileRestrictionsReplyMarkup(file *core.File, chats []*core.Chat) *tgbotapi.InlineKeyboardMarkup {
	keyboard := make([][]tgbotapi.InlineKeyboardButton, 0, len(chats)+1)

	for _, chat := range chats {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				addIsEnabledEmoji(
					chat.ID == file.Restriction.ChatID,
					chat.Title,
				),
				fmt.Sprintf(
					callbackFileRestrictionsChat,
					file.ID,
					chat.ID,
				),
			),
		))
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			textCommonBack,
			fmt.Sprintf("file:%d:refresh", file.ID),
		),
	))

	markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	return &markup
}

func (bot *Bot) onFileRestrictionsSetChatCBQ(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	fileID core.FileID,
	chatID core.ChatID,
) error {
	user := getUserCtx(ctx)

	result, err := bot.fileSrv.SetChatRestriction(
		ctx,
		user,
		fileID,
		chatID,
	)
	if err != nil {
		return errors.Wrap(err, "service set chat restriction")
	}

	chats, err := bot.chatSrv.GetChats(ctx, user)
	if err != nil {
		return errors.Wrap(err, "service query chats")
	}

	replyMarkup := bot.newFileRestrictionsReplyMarkup(
		result.File,
		chats,
	)

	go func() {
		if result.Disable {
			_ = bot.answerCallbackQuery(ctx, cbq, "Ограничение на загрузку отключено")
		} else {
			_ = bot.answerCallbackQuery(ctx, cbq, "Ограничение установлено")
		}
	}()

	return bot.send(ctx, tgbotapi.NewEditMessageReplyMarkup(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		*replyMarkup,
	))
}

func (bot *Bot) onFileRestrictionsChatCheck(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	fileID core.FileID,
) error {
	user := getUserCtx(ctx)

	status, err := bot.fileSrv.CheckFileRestrictionsChat(ctx, user, fileID)
	if err != nil {
		return errors.Wrap(err, "check file restrictions chat")
	}

	if !status.Ok {
		return bot.answerCallbackQueryAlert(ctx, cbq, "Я не наблюдаю тебе в подписчиках, подпишись чтобы получить доступ к файлу")
	}

	go func() {
		_ = bot.deleteMessage(ctx, cbq.Message)
	}()

	result, err := bot.fileSrv.RegisterDownload(ctx, user, status.File)
	if err != nil {
		return errors.Wrap(err, "register file download")
	}

	return bot.send(ctx, bot.renderNotOwnedFile(cbq.Message, result.File))
}
