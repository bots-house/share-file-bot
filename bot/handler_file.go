package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
	"github.com/lithammer/dedent"
)

const tgDomain = "t.me"

const (
	callbackFileRestrictions          = "file:%d:restrictions"
	callbackFileRestrictionsChat      = "file:%d:restrictions:chat-subscription:%d:toggl"
	callbackFileRestrictionsChatCheck = "file:%d:restrictions:chat:check"

	textButtonAbout = "Что это за бот?"
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

	var replyMarkup interface{}

	if file.LinkedPostURI.String != "" {
		replyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("« Пост в канале", file.LinkedPostURI.String),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Что это за бот?", cmdStart),
			),
		)
	} else {
		kb := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Что это за бот?"),
			),
		)

		kb.OneTimeKeyboard = true
		kb.ResizeKeyboard = true

		replyMarkup = kb
	}

	return bot.renderGenericFile(
		msg.Chat.ID,
		file.Kind,
		file.TelegramID,
		tg.EscapeMD(file.Caption.String),
		mdv2,
		replyMarkup,
	)
}

func (bot *Bot) renderOwnedFileCaption(file *service.OwnedFile) string {
	rows := []string{}

	if file.Caption.String != "" {

		rows = append(rows,
			"💬 __Описание__",
			"",
			tg.EscapeMD(file.Caption.String),
			"",
		)
	}

	rows = append(rows,
		"🔗 __Публичная ссылка__",
		"",
		fmt.Sprintf("https://%s/%s?start\\=%s",
			tg.EscapeMD(tgDomain),
			tg.EscapeMD(bot.client.Self.UserName),
			tg.EscapeMD(file.PublicID),
		),
		"",
	)

	if file.HasLinkedPostURI() && file.Restriction.HasChatID() {
		path, err := humanizePostURI(file.LinkedPostURI.String)
		if err == nil {
			rows = append(rows,
				fmt.Sprintf("Связанный пост: [%s](%s)",
					tg.EscapeMD(path),
					file.LinkedPostURI.String,
				),
				"",
			)
		}
	}

	rows = append(rows,
		"📈 __Статистика__",
		"",
	)

	rows = append(rows,
		fmt.Sprintf("*Загрузок*: `%d`", file.Stats.Total),
		fmt.Sprintf("*Уникальных загрузок*: `%d`", file.Stats.Unique),
		"",
	)

	if file.Restriction.HasChatID() {
		rows = append(rows,
			fmt.Sprintf("*Загрузок с подпиской*: `%d`", file.Stats.WithSubscription),
			fmt.Sprintf("*Загрузок с новой подпиской*: `%d`", file.Stats.NewSubscription),
			"",
		)
	}

	return strings.Join(rows, "\n")
}

func (bot *Bot) renderSubRequest(msg *tgbotapi.Message, sub *service.ChatSubRequest) tgbotapi.MessageConfig {
	var link string

	if sub.Username != "" {
		link = fmt.Sprintf("[@%s](https://t.me/%s)", tg.EscapeMD(sub.Username), sub.Username)
	} else {
		link = fmt.Sprintf("[%s](%s)", tg.EscapeMD(sub.Title), sub.JoinLink)
	}

	text := fmt.Sprintf(textFileSubRequest, link)

	out := tgbotapi.NewMessage(msg.Chat.ID, text)
	out.ParseMode = mdv2
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
		mdv2,
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

	if inputFile.Kind == core.KindAudio {
		_ = bot.sendText(ctx,
			user.ID,
			"⚠️ Упс, я не могу добавить этот файл, так как поддержка аудио временно отключена",
		)

		return nil
	}

	// delete user message for avoid trash in history
	go func() {
		_ = bot.deleteMessage(ctx, msg)
	}()

	file, err := bot.fileSrv.AddFile(ctx, user, inputFile)

	switch {
	case errors.Is(err, service.ErrUsersCantUploadFiles):
		_ = bot.sendText(ctx,
			user.ID,
			"✋ Загрузка файлов доступна только администраторам ботам",
		)

		return nil
	case err != nil:
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

	file, err := bot.fileSrv.GetFileByID(ctx, user, core.FileID(id))
	if errors.Is(err, core.ErrFileNotFound) {
		_ = bot.deleteMessage(ctx, cbq.Message)
		_ = bot.answerCallbackQueryAlert(ctx, cbq, "Файл был удален ранее")
		return nil, err
	} else if err != nil {
		return nil, errors.Wrap(err, "get file by id")
	}

	// user is not owner of file but try to access
	if file.OwnedFile == nil {
		if cbq != nil {
			_ = bot.answerCallbackQuery(ctx, cbq, "bad body, what you do?")
		}
		return nil, errors.New("can't manage file")
	}

	return file.OwnedFile, nil
}

func (bot *Bot) onFileRefreshCBQ(ctx context.Context, cbq *tgbotapi.CallbackQuery, id int) error {
	file, err := bot.getFileForOwner(ctx, cbq, id)
	if errors.Is(err, core.ErrFileNotFound) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "get file for owner")
	}

	caption := bot.renderOwnedFileCaption(file)

	edit := tgbotapi.NewEditMessageCaption(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		caption,
	)

	edit.ParseMode = mdv2
	replyMarkup := bot.renderOwnedFileReplyMarkup(file)
	edit.ReplyMarkup = &replyMarkup

	if err := bot.send(ctx, edit); err != nil {
		var tgErr *tgbotapi.Error

		if errors.As(err, &tgErr) {
			// TODO: use tg.IsErr...
			if strings.Contains(tgErr.Message, "message is not modified:") {
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
	if errors.Is(err, core.ErrFileNotFound) {
		return nil
	} else if err != nil {
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

	file, err := bot.getFileForOwner(ctx, cbq, id)
	if errors.Is(err, core.ErrFileNotFound) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "get file for owner")
	}

	if err := bot.fileSrv.DeleteFile(ctx, user, file.ID); err != nil {
		return errors.Wrap(err, "delete file")
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
	file, err := bot.getFileForOwner(ctx, cbq, fid)
	if errors.Is(err, core.ErrFileNotFound) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "get file for owner")
	}

	go func() {
		_ = bot.answerCallbackQuery(ctx, cbq, "")
	}()

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
		ParseMode: mdv2,
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
	switch {
	case errors.Is(err, service.ErrCantCheckMembership):
		//nolint:stylecheck
		return bot.answerCallbackQueryAlert(ctx, cbq, "🙅‍♂️ Я не могу выдать тебе файл, так как больше не являюсь админом канала на который требовалась подписка, свяжись с владельцем файла и передавай от меня привет!")
	case errors.Is(err, core.ErrFileNotFound):
		//nolint:stylecheck
		return bot.answerCallbackQueryAlert(ctx, cbq, "🙅‍♂️ Файл был удален владельцем")
	case err != nil:
		return errors.Wrap(err, "check file restrictions chat")
	}

	if !status.Ok {
		return bot.answerCallbackQueryAlert(ctx, cbq, "Я не наблюдаю тебя в подписчиках, подпишись чтобы получить доступ к файлу")
	}

	go func() {
		_ = bot.deleteMessage(ctx, cbq.Message)
		_ = bot.answerCallbackQuery(ctx, cbq, "🔓 Доступ к файлу получен")
	}()

	result, err := bot.fileSrv.RegisterDownload(ctx, user, status.File)
	if err != nil {
		return errors.Wrap(err, "register file download")
	}

	return bot.send(ctx, bot.renderNotOwnedFile(cbq.Message, result.File))
}

func (bot *Bot) onPublicFileHelp(ctx context.Context, cbq *tgbotapi.CallbackQuery) error {
	answer := tgbotapi.NewMessage(cbq.Message.Chat.ID, bot.getTextStart())
	answer.ParseMode = mdv2
	return bot.send(ctx, answer)
}
