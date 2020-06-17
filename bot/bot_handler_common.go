package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	imageHowToUpload = "https://telegra.ph/file/2de4c3f11a14eeb0adcfb.png"
	textHelp         = "Я помогу тебе поделится любым *документом* с подписчиками твоего канала. Отправь мне любой файл, а я в ответ дам тебе ссылку."
	textStart        = "Привет! 👋\n\n" + textHelp
	textNotDocument  = "На данный момент я работаю только с *документами* (*файлами*). Выбери нужный вариант при загрузке 👇"
)

func embeddWebPagePreview(txt string, link string) string {
	return fmt.Sprintf("[‎](%s)%s", link, txt)
}

func (bot *Bot) onHelp(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newAnswerMsg(ctx, msg, textHelp)
	return bot.send(ctx, answer)
}

func (bot *Bot) onStart(ctx context.Context, msg *tgbotapi.Message) error {
	answer := bot.newAnswerMsg(ctx, msg, textStart)
	return bot.send(ctx, answer)
}

func (bot *Bot) onNotDocument(ctx context.Context, msg *tgbotapi.Message) error {
	txt := embeddWebPagePreview(textNotDocument, imageHowToUpload)
	answer := bot.newAnswerMsg(ctx, msg, txt)
	return bot.send(ctx, answer)
}
