package bot

import (
	"context"
	"strings"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func escapeMarkdown(txt string) string {
	txt = strings.ReplaceAll(txt, "_", "\\_")
	txt = strings.ReplaceAll(txt, "*", "\\*")
	txt = strings.ReplaceAll(txt, "[", "\\[")
	txt = strings.ReplaceAll(txt, "`", "\\`")
	return txt
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

	result.ParseMode = tgbotapi.ModeMarkdown

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
