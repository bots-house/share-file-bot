package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const tgDomain = "t.me"

func (bot *Bot) renderNotOwnedFile(msg *tgbotapi.Message, doc *core.File) tgbotapi.DocumentConfig {
	share := tgbotapi.NewDocumentShare(int64(msg.From.ID), doc.TelegramID)
	share.ParseMode = tgbotapi.ModeMarkdown
	share.Caption = escapeMarkdown(doc.Caption.String)
	return share
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
	)
}

func (bot *Bot) renderOwnedFile(msg *tgbotapi.Message, doc *service.OwnedFile) tgbotapi.DocumentConfig {
	share := tgbotapi.NewDocumentShare(int64(msg.From.ID), doc.TelegramID)

	share.ParseMode = tgbotapi.ModeMarkdown
	share.Caption = bot.renderOwnedFileCaption(doc)
	share.ReplyMarkup = bot.renderOwnedFileReplyMarkup(doc)

	return share
}

func (bot *Bot) onFile(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)

	go func() {
		if err := bot.send(ctx, tgbotapi.NewDeleteMessage(
			msg.Chat.ID,
			msg.MessageID,
		)); err != nil {
			log.Warn(ctx, "can't delete incoming message", "chat_id", msg.Chat.ID, "msg_id", msg.MessageID)
		}
	}()

	// spew.Dump(msg)
	doc, err := bot.fileSrv.AddFile(ctx, user, &service.InputFile{
		FileID:   msg.Document.FileID,
		Caption:  msg.Caption,
		MIMEType: msg.Document.MimeType,
		Size:     msg.Document.FileSize,
		Name:     msg.Document.FileName,
	})

	if err != nil {
		return errors.Wrap(err, "service add file")
	}

	result := bot.renderOwnedFile(msg, doc)

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

	if err := bot.send(ctx, tgbotapi.NewDeleteMessage(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
	)); err != nil {
		log.Warn(ctx, "can't delete message")
	}

	return bot.answerCallbackQuery(ctx, cbq, "✅ Документ удален")
}
