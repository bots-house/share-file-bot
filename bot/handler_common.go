package bot

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lithammer/dedent"
)

const (
	textUnsupportedFileKind = "К сожалению, я не поддерживаю данный тип файлов. На данный момент я умею работать только с документами, видео, фото, аудио и голосовыми. Отправь и перешли мне сообщение перечисленного типа, а в ответ я дам тебе ссылку."
	mdv2                    = "MarkdownV2"
)

var (
	textHelp = dedent.Dedent(`
		Я помогу тебе поделится любым медиафайлом (фото, видео, документы, аудио, голосовые) с подписчиками твоего канала. 
		Отправь любой из перечисленных файлов, а я в ответ дам тебе ссылку. Желательно указать подпись, чтобы человек не забыл кто ему это пошарил.
		Так же ты можешь подключить свой канал или чат и установить ограничение на доступ к медиафайлу только своим подписчикам. 

		/settings - для более тонкой настройки

		Поддержка: @share\_file\_support
	`)

	textStart = "Привет! 👋\n" + textHelp
)

func (bot *Bot) onHelp(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newAnswerMsg(msg, textHelp)
	return bot.send(ctx, answer)
}

func (bot *Bot) onStart(ctx context.Context, msg *tgbotapi.Message) error {
	if args := msg.CommandArguments(); args != "" {
		user := getUserCtx(ctx)

		log.Debug(ctx, "query file", "public_id", args)
		result, err := bot.fileSrv.GetFileByPublicID(ctx, user, args)
		if errors.Cause(err) == core.ErrFileNotFound {
			answer := bot.newAnswerMsg(msg, "😐Ничего не знаю о таком файле, проверь ссылку...")
			return bot.send(ctx, answer)
		} else if err != nil {
			return errors.Wrap(err, "download file")
		}

		switch {
		case result.OwnedFile != nil:
			return bot.send(ctx, bot.renderOwnedFile(msg, result.OwnedFile))
		case result.File != nil:
			return bot.send(ctx, bot.renderNotOwnedFile(msg, result.File))
		case result.ChatSubRequest != nil:
			return bot.send(ctx, bot.renderSubRequest(msg, result.ChatSubRequest))
		default:
			log.Error(ctx, "bad result")
		}
	}

	answer := bot.newAnswerMsg(msg, textStart)
	return bot.send(ctx, answer)
}

func (bot *Bot) onUnsupportedFileKind(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newReplyMsg(msg, textUnsupportedFileKind)
	return bot.send(ctx, answer)
}

func (bot *Bot) onVersion(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newReplyMsg(msg, "`"+bot.revision+"`")
	return bot.send(ctx, answer)
}
