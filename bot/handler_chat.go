package bot

import (
	"context"

	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/snip"
	"github.com/bots-house/share-file-bot/service"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
)

func (bot *Bot) onChatNewTitle(ctx context.Context, msg *tgbotapi.Message) error {
	if err := bot.chatSrv.UpdateTitle(ctx, msg.Chat.ID, msg.NewChatTitle); err != nil {
		return errors.Wrap(err, "update title")
	}

	return nil
}

func (bot *Bot) onChatNewPost(ctx context.Context, post *tgbotapi.Message) error {
	urls := parseURLsFromChannelPost(post)

	if len(urls) == 0 {
		return nil
	}

	urls = snip.UniqueizeStrings(urls)

	log.Info(ctx, "process channel post uries", "uries_count", len(urls), "chat_id", post.Chat.ID, "message_id", post.MessageID)
	if err := bot.chatSrv.ProcessChannelPostURIes(ctx, &service.ChannelPostInfo{
		ChatID:       post.Chat.ID,
		ChatUsername: post.Chat.UserName,
		PostID:       post.MessageID,
	}, urls); err != nil {
		return errors.Wrap(err, "process channel post uries")
	}

	return nil
}
