package bot

import (
	"context"
	"net/url"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
)

func getURLsFromMessageEntities(entities *[]tgbotapi.MessageEntity) []string {
	if entities == nil {
		return []string{}
	}

	result := make([]string, 0, len(*entities))

	for _, entity := range *entities {
		if entity.Type == "text_link" {
			result = append(result, entity.URL)
		}
	}

	return result
}

func getURLsFromMessageReplyMarkup(rm *tgbotapi.InlineKeyboardMarkup) []string {
	if rm == nil {
		return []string{}
	}

	result := make([]string, 0, len(rm.InlineKeyboard))
	for _, row := range rm.InlineKeyboard {
		for _, btn := range row {
			if btn.URL != nil {
				result = append(result, *btn.URL)
			}
		}
	}

	return result
}

func (bot *Bot) send(_ context.Context, s tgbotapi.Chattable) error {
	// spew.Dump(msg)
	_, err := bot.client.Send(s)
	return err
}

func (bot *Bot) sendText(ctx context.Context, uid core.UserID, text string) error {
	return bot.send(ctx, tgbotapi.NewMessage(int64(uid), text))
}

func (bot *Bot) newAnswerMsg(msg *tgbotapi.Message, text string) *tgbotapi.MessageConfig {
	result := tgbotapi.NewMessage(
		int64(msg.From.ID),
		text,
	)

	result.ParseMode = mdv2

	return &result
}

func (bot *Bot) newReplyMsg(msg *tgbotapi.Message, text string) *tgbotapi.MessageConfig {
	result := bot.newAnswerMsg(msg, text)
	result.ReplyToMessageID = msg.MessageID
	return result
}

func (bot *Bot) answerCallbackQuery(_ context.Context, cbq *tgbotapi.CallbackQuery, text string) error {
	_, err := bot.client.AnswerCallbackQuery(tgbotapi.NewCallback(
		cbq.ID,
		text,
	))

	return err
}

func (bot *Bot) answerCallbackQueryAlert(_ context.Context, cbq *tgbotapi.CallbackQuery, text string) error {
	answ := tgbotapi.NewCallback(
		cbq.ID,
		text,
	)

	answ.ShowAlert = true

	_, err := bot.client.AnswerCallbackQuery(answ)

	return err
}

func (bot *Bot) detectKind(msg *tgbotapi.Message) core.Kind {
	switch {
	case msg.Animation != nil:
		return core.KindAnimation
	case msg.Audio != nil:
		return core.KindAudio
	case msg.Photo != nil:
		return core.KindPhoto
	case msg.Video != nil:
		return core.KindVideo
	case msg.Voice != nil:
		return core.KindVoice
	case msg.Document != nil:
		return core.KindDocument
	default:
		return core.KindUnknown
	}
}

func (bot *Bot) extractInputFileFromMessage(msg *tgbotapi.Message) *service.InputFile {
	switch kind := bot.detectKind(msg); kind {
	case core.KindDocument:
		return &service.InputFile{
			FileID:   msg.Document.FileID,
			Caption:  msg.Caption,
			Kind:     kind,
			MIMEType: msg.Document.MimeType,
			Size:     msg.Document.FileSize,
			Name:     msg.Document.FileName,
		}
	case core.KindAnimation:
		return &service.InputFile{
			FileID:   msg.Animation.FileID,
			Caption:  msg.Caption,
			Kind:     kind,
			MIMEType: msg.Animation.MimeType,
			Size:     msg.Animation.FileSize,
		}
	case core.KindAudio:
		return &service.InputFile{
			FileID:   msg.Audio.FileID,
			Caption:  msg.Caption,
			Kind:     core.KindAudio,
			Metadata: core.NewMetadataAudio(msg.Audio.Title, msg.Audio.Performer),
			MIMEType: msg.Audio.MimeType,
			Size:     msg.Audio.FileSize,
		}
	case core.KindPhoto:
		total := len(*msg.Photo)

		return &service.InputFile{
			FileID:  (*msg.Photo)[total-1].FileID,
			Caption: msg.Caption,
			Kind:    core.KindPhoto,
			Size:    (*msg.Photo)[total-1].FileSize,
		}
	case core.KindVideo:
		return &service.InputFile{
			FileID:   msg.Video.FileID,
			Caption:  msg.Caption,
			Kind:     core.KindVideo,
			Size:     msg.Video.FileSize,
			MIMEType: msg.Video.MimeType,
		}
	case core.KindVoice:
		return &service.InputFile{
			FileID:   msg.Voice.FileID,
			Caption:  msg.Caption,
			Kind:     core.KindVoice,
			Size:     msg.Voice.FileSize,
			MIMEType: msg.Voice.MimeType,
		}
	default:
		return nil
	}
}

func humanizePostURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, "parse uri")
	}

	q := u.Query()

	var chat string

	if _, ok := q["domain"]; ok {
		chat = q.Get("domain")
	} else if _, ok := q["channel"]; ok {
		chat = q.Get("channel")
	} else {
		return "", errors.New("chat not found")
	}

	postID := q.Get("post")

	return chat + "/" + postID, nil
}
