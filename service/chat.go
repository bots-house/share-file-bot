package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/store"
	"github.com/friendsofgo/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Chat struct {
	Telegram *tgbotapi.BotAPI
	Txier    store.Txier
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
	ErrChatNotFoundOrBotIsNotAdmin = errors.New("chat not found or bot is not admin")
	ErrChatIsUser                  = errors.New("chat is private (user)")
	ErrBotIsNotChatAdmin           = errors.New("bot is not admin")
	ErrUserIsNotChatAdmin          = errors.New("user is not admin")
)

func (srv *Chat) UpdateTitle(ctx context.Context, chatID int64, title string) error {
	return srv.Txier(ctx, func(ctx context.Context) error {
		chats, err := srv.Chat.Query().TelegramID(chatID).All(ctx)
		if err != nil {
			return errors.Wrap(err, "query chats")
		}

		for _, chat := range chats {
			updated := chat.Patch(func(chat *core.Chat) {
				chat.Title = title
			})

			if !updated {
				continue
			}

			log.Info(ctx, "update chat title", "id", chatID, "title", title)
			if err := srv.Chat.Update(ctx, chat); err != nil {
				return errors.Wrap(err, "update fail")
			}
		}

		return nil
	})

}

// Add links chat to Share File Bot.
func (srv *Chat) Add(ctx context.Context, user *core.User, identity ChatIdentity) (*FullChat, error) {
	chatInfo, err := srv.Telegram.GetChat(tgbotapi.ChatConfig{
		ChatID:             identity.ID,
		SuperGroupUsername: identity.Username,
	})

	if tg.IsChatNotFoundError(err) {
		return nil, ErrChatNotFoundOrBotIsNotAdmin
	} else if err != nil {
		return nil, errors.Wrap(err, "get chat")
	}

	typ, err := srv.getTypeFromChatInfo(&chatInfo)
	if err != nil {
		return nil, errors.Wrap(err, "get type from chat info")
	}

	admins, err := srv.Telegram.GetChatAdministrators(tgbotapi.ChatConfig{
		ChatID:             identity.ID,
		SuperGroupUsername: identity.Username,
	})

	if tg.IsMemberListIsInaccessible(err) {
		return nil, ErrBotIsNotChatAdmin
	} else if err != nil {
		return nil, errors.Wrap(err, "get chat admins")
	}

	if !srv.isUserAdmin(admins, srv.Telegram.Self.ID) {
		return nil, ErrBotIsNotChatAdmin
	}

	if !srv.isUserAdmin(admins, int(user.ID)) {
		return nil, ErrUserIsNotChatAdmin
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

	return &FullChat{chat}, nil
}

func (srv *Chat) isUserAdmin(admins []tgbotapi.ChatMember, userID int) bool {
	for _, admin := range admins {
		if admin.User.ID == userID {
			return admin.IsAdministrator() || admin.IsCreator()
		}
	}
	return false
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
		return core.ChatType(0), errors.New("unknown chat type")
	}
}

// GetChats returns chats of user
func (srv *Chat) GetChats(ctx context.Context, user *core.User) ([]*core.Chat, error) {
	chats, err := srv.Chat.Query().OwnerID(user.ID).All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query user chats")
	}
	return chats, nil
}

type FullChat struct {
	*core.Chat
}

func (srv *Chat) GetChat(ctx context.Context, user *core.User, id core.ChatID) (*FullChat, error) {
	chat, err := srv.Chat.Query().OwnerID(user.ID).ID(id).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query chat")
	}
	return &FullChat{chat}, nil
}

// DisconnectChat disconnect user linked chat and optionaly leave it.
func (srv *Chat) DisconnectChat(
	ctx context.Context,
	user *core.User,
	id core.ChatID,
	leave bool,
) error {
	return srv.Txier(ctx, func(ctx context.Context) error {
		return srv.disconnectChat(ctx, user, id, false)
	})
}

func (srv *Chat) disconnectChat(
	ctx context.Context,
	user *core.User,
	id core.ChatID,
	leave bool,
) error {
	chat, err := srv.Chat.Query().OwnerID(user.ID).ID(id).One(ctx)
	if err != nil {
		return errors.Wrap(err, "query chat")
	}

	count, err := srv.Chat.Query().ID(chat.ID).Delete(ctx)
	if err != nil {
		return errors.Wrap(err, "delete chats")
	}

	if count == 0 {
		return core.ErrChatNotFound
	} else if count > 1 {
		return store.ErrTooManyAffectedRows
	}

	if leave {
		_, err := srv.Telegram.LeaveChat(tgbotapi.ChatConfig{
			ChatID: chat.TelegramID,
		})

		if err != nil {
			log.Warn(ctx, "can't leave chat", "chat_id", chat.TelegramID, "err", err)
		}
	}

	return nil
}
