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
		escapeMarkdown(doc.SecretID),
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

func (bot *Bot) answerCallbackQuery(ctx context.Context, cbq *tgbotapi.CallbackQuery, text string) error {
	_, err := bot.client.AnswerCallbackQuery(tgbotapi.NewCallback(
		cbq.ID,
		text,
	))

	return err
}

func (bot *Bot) onDocumentRefreshCBQ(ctx context.Context, cbq *tgbotapi.CallbackQuery, id int) error {
	user := getUserCtx(ctx)

	doc, err := bot.docSrv.GetDocumentByID(ctx, user, core.DocumentID(id))
	if err != nil {
		return errors.Wrap(err, "get document by id")
	}

	// user is not owner of document but try to access
	if doc.OwnedDocument == nil {
		return bot.answerCallbackQuery(ctx, cbq, "bad body, what you do?")
	}

	caption := bot.renderOwnedDocumentCaption(doc.OwnedDocument)

	edit := tgbotapi.NewEditMessageCaption(
		cbq.Message.Chat.ID,
		cbq.Message.MessageID,
		caption,
	)

	edit.ParseMode = tgbotapi.ModeMarkdown
	replyMarkup := bot.renderOwnedDocumentReplyMarkup(doc.OwnedDocument)
	edit.ReplyMarkup = &replyMarkup

	if err := bot.send(ctx, edit); err != nil {
		if err, ok := err.(tgbotapi.Error); ok {
			if strings.Contains(err.Message, "message is not modified:") {
				return bot.answerCallbackQuery(ctx, cbq, "🤷 Ничего не изменилось")
			}
		}
		return errors.Wrap(err, "edit message error")
	}

	return bot.answerCallbackQuery(ctx, cbq, "Обновлено!")
}
