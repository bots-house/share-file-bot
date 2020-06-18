package bot

import (
	"context"
	"strconv"

	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

func (bot *Bot) answerInlineQuery(ctx context.Context, iq *tgbotapi.InlineQuery, answer tgbotapi.InlineConfig) error {
	answer.InlineQueryID = iq.ID
	_, err := bot.client.AnswerInlineQuery(answer)
	return err
}

type inlineQueryResultDocument struct {
	Type           string `json:"type"`             // required
	ID             string `json:"id"`               // required
	Title          string `json:"title"`            // required
	DocumentFileID string `json:"document_file_id"` // required
	Description    string `json:"description,omitempty"`
	Caption        string `json:"caption,omitempty"`
	ParseMode      string `json:"parse_mode,omitempty"`
}

func (bot *Bot) onSearch(ctx context.Context, iq *tgbotapi.InlineQuery) error {
	user := getUserCtx(ctx)

	var offset int

	if iq.Offset != "" {
		var err error
		offset, err = strconv.Atoi(iq.Offset)
		if err != nil {
			return errors.Wrap(err, "parse offset")
		}
	}

	query := &service.SearchQuery{
		Query:  iq.Query,
		Offset: offset,
	}

	ctx = boil.WithDebug(ctx, true)

	result, err := bot.docSrv.Search(
		ctx,
		user,
		query,
	)
	if err != nil {
		return errors.Wrap(err, "search docs")
	}

	items := make([]interface{}, len(result.Items))

	for i, doc := range result.Items {
		items[i] = &inlineQueryResultDocument{
			Type:           "document",
			ID:             strconv.Itoa(int(doc.ID)),
			Title:          doc.Name,
			DocumentFileID: doc.FileID,
			// Description:    doc.Caption.String,
		}
	}

	answer := tgbotapi.InlineConfig{
		Results: items,
	}
	return bot.answerInlineQuery(ctx, iq, answer)
}
