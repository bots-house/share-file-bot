package bot

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	imageHowToUpload = "https://telegra.ph/file/2de4c3f11a14eeb0adcfb.png"
	textHelp         = "Я помогу тебе поделится любым *документом* с подписчиками твоего канала. Отправь мне любой файл, а я в ответ дам тебе ссылку. Так же рекомендую указать подпись к файлу, чтобы человек не забыл кто ему этот файл пошарил 🤗"
	textStart        = "Привет! 👋\n\n" + textHelp
	textNotDocument  = "На данный момент я работаю только с *документами* (*файлами*). Выбери нужный вариант при загрузке 👇"
)

func (bot *Bot) onHelp(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newAnswerMsg(ctx, msg, textHelp)
	return bot.send(ctx, answer)
}

func (bot *Bot) onStart(ctx context.Context, msg *tgbotapi.Message) error {
	if args := msg.CommandArguments(); args != "" {
		user := getUserCtx(ctx)

		log.Debug(ctx, "query document", "public_id", args)
		result, err := bot.docSrv.GetDocumentByPublicID(ctx, user, args)
		if errors.Cause(err) == core.ErrDocumentNotFound {
			answer := bot.newAnswerMsg(ctx, msg, "😐Ничего не знаю о таком файле, проверь ссылку...")
			return bot.send(ctx, answer)
		} else if err != nil {
			return errors.Wrap(err, "download document")
		}

		if result.OwnedDocument != nil {
			return bot.send(ctx, bot.renderOwnedDocument(msg, result.OwnedDocument))
		} else if result.Document != nil {
			return bot.send(ctx, bot.renderNotOwnedDocument(msg, result.Document))
		} else {
			log.Error(ctx, "bad result")
		}
	}

	answer := bot.newAnswerMsg(ctx, msg, textStart)
	return bot.send(ctx, answer)
}

func (bot *Bot) onNotDocument(ctx context.Context, msg *tgbotapi.Message) error {
	txt := embeddWebPagePreview(textNotDocument, imageHowToUpload)
	answer := bot.newAnswerMsg(ctx, msg, txt)
	return bot.send(ctx, answer)
}
