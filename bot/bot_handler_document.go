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

func (bot *Bot) renderNotOwnedDocument(msg *tgbotapi.Message, doc *core.Document) tgbotapi.DocumentConfig {
	share := tgbotapi.NewDocumentShare(int64(msg.From.ID), doc.FileID)
	share.ParseMode = tgbotapi.ModeMarkdown
	share.Caption = escapeMarkdown(doc.Caption.String)
	return share
}

func (bot *Bot) renderOwnedDocumentCaption(doc *service.OwnedDocument) string {
	rows := []string{}

	if doc.Caption.String != "" {
		rows = append(rows,
			fmt.Sprintf("*Описание*: %s", escapeMarkdown(doc.Caption.String)),
			"",
		)
	}

	rows = append(rows,
		fmt.Sprintf("*Кол-во загрузок*: `%d`", doc.Stats.Total),
		fmt.Sprintf("*Кол-во уникальных загрузок*: `%d`", doc.Stats.Unique),
		"",
	)

	rows = append(rows, fmt.Sprintf("*Публичная ссылка*: https://%s/%s?start=%s",
		tgDomain,
		escapeMarkdown(bot.client.Self.UserName),
		escapeMarkdown(doc.PublicID),
	))

	return strings.Join(rows, "\n")
}

func (bot *Bot) renderOwnedDocumentReplyMarkup(doc *service.OwnedDocument) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Обновить",
				fmt.Sprintf("document:%d:refresh", doc.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Удалить",
				fmt.Sprintf("document:%d:delete", doc.ID),
			),
		),
	)
}

func (bot *Bot) renderOwnedDocument(msg *tgbotapi.Message, doc *service.OwnedDocument) tgbotapi.DocumentConfig {
	share := tgbotapi.NewDocumentShare(int64(msg.From.ID), doc.FileID)

	share.ParseMode = tgbotapi.ModeMarkdown
	share.Caption = bot.renderOwnedDocumentCaption(doc)
	share.ReplyMarkup = bot.renderOwnedDocumentReplyMarkup(doc)

	return share
}

func (bot *Bot) onDocument(ctx context.Context, msg *tgbotapi.Message) error {
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
	doc, err := bot.docSrv.AddDocument(ctx, user, &service.InputDocument{
		FileID:   msg.Document.FileID,
		Caption:  msg.Caption,
		MIMEType: msg.Document.MimeType,
		Size:     msg.Document.FileSize,
		Name:     msg.Document.FileName,
	})

	if err != nil {
		return errors.Wrap(err, "service add document")
	}

	result := bot.renderOwnedDocument(msg, doc)

	return bot.send(ctx, result)
}

func (bot *Bot) getDocumentForOwner(ctx context.Context, cbq *tgbotapi.CallbackQuery, id int) (*service.OwnedDocument, error) {
	user := getUserCtx(ctx)

	doc, err := bot.docSrv.GetDocumentByID(ctx, user, core.DocumentID(id))
	if err != nil {
		return nil, errors.Wrap(err, "get document by id")
	}

	// user is not owner of document but try to access
	if doc.OwnedDocument == nil {
		if cbq != nil {
			_ = bot.answerCallbackQuery(ctx, cbq, "bad body, what you do?")
		}
		return nil, errors.New("can't manage document")
	}

	return doc.OwnedDocument, nil
}

func (bot *Bot) onDocumentRefreshCBQ(ctx context.Context, cbq *tgbotapi.CallbackQuery, id int) error {
	doc, err := bot.getDocumentForOwner(ctx, cbq, id)
	if err != nil {
		return errors.Wrap(err, "get document for owner")
	}

	caption := bot.renderOwnedDocumentCaption(doc)

	edit := tgbotapi.NewEditMessageCaption(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		caption,
	)

	edit.ParseMode = tgbotapi.ModeMarkdown
	replyMarkup := bot.renderOwnedDocumentReplyMarkup(doc)
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

func (bot *Bot) onDocumentDeleteCBQ(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	id int,
) error {
	doc, err := bot.getDocumentForOwner(ctx, cbq, id)
	if err != nil {
		return errors.Wrap(err, "get document for owner")
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
				fmt.Sprintf("document:%d:delete:confirm", doc.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Нет",
				fmt.Sprintf("document:%d:refresh", doc.ID),
			),
		),
	)

	edit.ReplyMarkup = &markup

	return bot.send(ctx, edit)
}

func (bot *Bot) onDocumentDeleteConfirmCBQ(
	ctx context.Context,
	cbq *tgbotapi.CallbackQuery,
	id int,
) error {
	user := getUserCtx(ctx)
	docID := core.DocumentID(id)

	if err := bot.docSrv.DeleteDocument(ctx, user, docID); err == core.ErrDocumentNotFound {
		return bot.answerCallbackQuery(ctx, cbq, "Файл не найден")
	} else if err != nil {
		return errors.Wrap(err, "service delete document")
	}

	if err := bot.send(ctx, tgbotapi.NewDeleteMessage(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
	)); err != nil {
		log.Warn(ctx, "can't delete message")
	}

	return bot.answerCallbackQuery(ctx, cbq, "✅ Документ удален")
}
