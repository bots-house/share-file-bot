package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/core"
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

func (bot *Bot) renderOwnedDocument(msg *tgbotapi.Message, doc *service.OwnedDocument) tgbotapi.DocumentConfig {
	share := tgbotapi.NewDocumentShare(int64(msg.From.ID), doc.FileID)
	share.ParseMode = tgbotapi.ModeMarkdown

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

	share.Caption = strings.Join(rows, "\n")

	return share
}

func (bot *Bot) onDocument(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)
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
