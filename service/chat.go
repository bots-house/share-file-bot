package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Chat struct {
	Telegram *tgbotapi.BotAPI
	Chat     core.ChatStore
}

type ChatIdentity struct {
	ID       int64
	Username string
}

func NewChatIdentityFromID(id int64) ChatIdentity {
	return ChatIdentity{ID: id}
}

func NewChatIdentityFromUsername(un string) ChatIdentity {
	return ChatIdentity{Username: un}
}

var (
	ErrBotIsNotChatAdmin  = errors.New("bot is not admin")
	ErrUserIsNotChatAdmin = errors.New("user is not admin")
)

// Add links chat to Share File Bot.
func (srv *Chat) Add(ctx context.Context, user *core.User, identity ChatIdentity) (*core.Chat, error) {
	chatInfo, err := srv.Telegram.GetChat(tgbotapi.ChatConfig{
		ChatID:             identity.ID,
		SuperGroupUsername: identity.Username,
	})

	if err != nil {
		return nil, errors.Wrap(err, "get chat")
	}

	typ, err := srv.getTypeFromChatInfo(&chatInfo)
	if err != nil {
		return nil, errors.Wrap(err, "get type from chat info")
	}

	chat := core.NewChat(
		chatInfo.ID,
		chatInfo.Title,
		typ,
		user.ID,
	)

	if err := srv.Chat.Add(ctx, chat); err != nil {
		return nil, errors.Wrap(err, "add chat to store")
	}

	return chat, nil
}

func (srv *Chat) getTypeFromChatInfo(info *tgbotapi.Chat) (core.ChatType, error) {
	switch {
	case info.IsChannel():
		return core.ChatTypeChannel, nil
	case info.IsGroup():
		return core.ChatTypeGroup, nil
	case info.IsSuperGroup():
		return core.ChatTypeSuperGroup, nil
	default:
		return core.ChatType(0), errors.New("unkown chat type")
	}
}

func (srv *Chat) checkUserIsAdmin(ctx context.Context, identity ChatIdentity, userID int) error {
	member, err := srv.Telegram.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID:             identity.ID,
		SuperGroupUsername: identity.Username,
		UserID:             userID,
	})

	if err != nil {
		return errors.Wrap(err, "get chat member")
	}

	if !(member.IsAdministrator() || member.IsCreator()) {
		return ErrUserIsNotChatAdmin
	}

	return nil
}

func (srv *Chat) checkBotIsAdmin(ctx context.Context, identity ChatIdentity) error {
	member, err := srv.Telegram.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID:             identity.ID,
		SuperGroupUsername: identity.Username,
		UserID:             srv.Telegram.Self.ID,
	})

	if err != nil {
		return errors.Wrap(err, "get chat member")
	}

	if !member.IsAdministrator() {
		return ErrBotIsNotChatAdmin
	}

	return nil
}
