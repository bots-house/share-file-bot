package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bots-house/share-file-bot/bot/state"
	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
	"github.com/lithammer/dedent"
)

const (
	textUnsupportedFileKind = "К сожалению, я не поддерживаю данный тип файлов\\. На данный момент я умею работать только с документами, видео, фото, аудио и голосовыми\\. Отправь и перешли мне сообщение перечисленного типа, а в ответ я дам тебе ссылку\\."
	mdv2                    = "MarkdownV2"
)

var (
	textHello = "Привет\\! 👋\n"

	textHelp = dedent.Dedent(`
		Я помогу тебе поделиться любым медиафайлом \(фото, видео, документы, аудио, голосовые\) с подписчиками твоего канала\. 
		Отправь любой из перечисленных файлов, а я в ответ дам тебе ссылку\. Желательно указать подпись, чтобы человек не забыл кто ему это пошарил\.
		Так же ты можешь подключить свой канал или чат и установить ограничение на доступ к медиафайлу только своим подписчикам\. 

		/settings \- для более тонкой настройки

		Поддержка: @share\_file\_support
		Новости и обновления: @share\_file\_news
	`)

	textStart = textHello + textHelp
)

func (bot *Bot) getTextStart() string {
	if bot.textHelp != "" {
		return textHello + bot.textHelp
	}

	return textStart
}

func (bot *Bot) getTextHelp() string {
	if bot.textHelp != "" {
		return bot.textHelp
	}

	return textHelp
}

func (bot *Bot) onHelp(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newAnswerMsg(msg, bot.getTextHelp())
	answer.ParseMode = mdv2
	return bot.send(ctx, answer)
}

func (bot *Bot) onStart(ctx context.Context, msg *tgbotapi.Message) error {
	user := getUserCtx(ctx)

	// reset state
	if err := bot.state.Set(ctx, user.ID, state.Empty); err != nil {
		return errors.Wrap(err, "update state")
	}

	if args := msg.CommandArguments(); args != "" {
		log.Debug(ctx, "query file", "public_id", args)
		result, err := bot.fileSrv.GetFileByPublicID(ctx, user, args)

		switch {
		case errors.Is(err, core.ErrFileNotFound):
			answer := bot.newAnswerMsg(msg, "😐 Ничего не знаю о таком файле, проверь ссылку...")
			return bot.send(ctx, answer)
		case errors.Is(err, service.ErrFileViolatesCopyright):
			answer := bot.newAnswerMsg(msg, "😐 К сожалению, на данный файл поступила жалоба от правообладателей и мы были вынужденны его удалить.")
			return bot.send(ctx, answer)
		case errors.Is(err, service.ErrCantCheckMembership):
			//nolint:stylecheck
			answer := bot.newAnswerMsg(msg, "🙅‍♂️‍ Я не могу выдать тебе файл, так как больше не являюсь админом канала на который требовалась подписка, свяжись с владельцем файла и передавай от меня привет!")
			return bot.send(ctx, answer)
		case err != nil:
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

	answer := bot.newAnswerMsg(msg, bot.getTextStart())
	return bot.send(ctx, answer)
}

func (bot *Bot) onUnsupportedFileKind(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newReplyMsg(msg, textUnsupportedFileKind)
	return bot.send(ctx, answer)
}

func (bot *Bot) onVersion(ctx context.Context, msg *tgbotapi.Message) error {
	text := strings.Join([]string{
		fmt.Sprintf("**Version**: `%s`", tg.EscapeMD(bot.buildInfo.Version)),
		fmt.Sprintf("**Ref**: `%s`", tg.EscapeMD(bot.buildInfo.Ref)),
		fmt.Sprintf("**Time**: `%s`", tg.EscapeMD(bot.buildInfo.Time)),
	}, "\n")

	answer := bot.newReplyMsg(msg, text)

	return bot.send(ctx, answer)
}
