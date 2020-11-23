package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/snip"
	"github.com/bots-house/share-file-bot/pkg/tg"
	"github.com/bots-house/share-file-bot/store"
	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/friendsofgo/errors"
	"golang.org/x/sync/errgroup"
)

type Chat struct {
	Telegram *tgbotapi.BotAPI
	Txier    store.Txier

	File     core.FileStore
	Chat     core.ChatStore
	Download core.DownloadStore
}

type ChatIdentity struct {
	ID       int64
	Username string
}

func NewChatIdentityFromID(id int64) ChatIdentity {
	return ChatIdentity{ID: id}
}

func NewChatIdentityFromUsername(un string) ChatIdentity {
	if !strings.HasPrefix(un, "@") {
		un = "@" + un
	}
	return ChatIdentity{Username: un}
}

var (
	ErrChatNotFoundOrBotIsNotAdmin = errors.New("chat not found or bot is not admin")
	ErrChatIsUser                  = errors.New("chat is private (user)")
	ErrBotIsNotChatAdmin           = errors.New("bot is not admin")
	ErrBotNotEnoughRights          = errors.New("bot not has rights")
	ErrUserIsNotChatAdmin          = errors.New("user is not admin")
	ErrChatAlreadyConnected        = errors.New("chat already connected")
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

	if tg.IsChatNotFoundError(err) || tg.IsBotIsNotMember(err) {
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

	if tg.IsMemberListIsInaccessible(err) || tg.IsBotIsNotMember(err) {
		return nil, ErrBotIsNotChatAdmin
	} else if err != nil {
		return nil, errors.Wrap(err, "get chat admins")
	}

	if !srv.isUserAdmin(admins, srv.Telegram.Self.ID, nil) {
		return nil, ErrBotIsNotChatAdmin
	}

	if !srv.isUserAdmin(admins, srv.Telegram.Self.ID, func(member tgbotapi.ChatMember) bool {
		return member.CanInviteUsers
	}) {
		return nil, ErrBotNotEnoughRights
	}

	if !srv.isUserAdmin(admins, int(user.ID), nil) {
		return nil, ErrUserIsNotChatAdmin
	}

	_, err = srv.Telegram.GetInviteLink(tgbotapi.ChatConfig{
		ChatID: chatInfo.ID,
	})

	if err != nil {
		return nil, errors.Wrap(err, "can't get invite link")
	}

	chat := core.NewChat(
		chatInfo.ID,
		chatInfo.Title,
		typ,
		user.ID,
	)

	if err := srv.Chat.Add(ctx, chat); err != nil {
		if err == ErrChatAlreadyConnected {
			return nil, ErrChatAlreadyConnected
		}
		return nil, errors.Wrap(err, "add chat to store")
	}

	return &FullChat{Chat: chat}, nil
}

func (srv *Chat) isUserAdmin(admins []tgbotapi.ChatMember, userID int, rights func(m tgbotapi.ChatMember) bool) bool {
	for _, admin := range admins {
		if admin.User.ID == userID {
			base := admin.IsAdministrator() || admin.IsCreator()

			if rights != nil {
				base = base && rights(admin)
			}

			return base
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

	// Count of files
	Files int

	// Stats of downloads
	Stats *core.ChatDownloadStats
}

func (chat *FullChat) GetStats() *core.ChatDownloadStats {
	if chat.Stats == nil {
		return &core.ChatDownloadStats{}
	}

	return chat.Stats
}

func (srv *Chat) GetChat(ctx context.Context, user *core.User, id core.ChatID) (*FullChat, error) {
	chat, err := srv.Chat.Query().OwnerID(user.ID).ID(id).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query chat")
	}

	full := &FullChat{Chat: chat}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		filesCount, err := srv.File.Query().
			RestrictionChatID(chat.ID).
			Count(ctx)

		if err != nil {
			return errors.Wrap(err, "query files count")
		}

		full.Files = filesCount

		return nil
	})

	g.Go(func() error {
		stats, err := srv.Download.GetChatStats(ctx, chat.ID)
		if err != nil {
			return errors.Wrap(err, "query chat stats")
		}

		full.Stats = stats

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "collect stats")
	}

	return full, nil
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

type ChannelPostInfo struct {
	ChatID       int64
	ChatUsername string
	PostID       int
}

// Link returns tg:// deeplink to post
func (info *ChannelPostInfo) Link() string {
	args := url.Values{}

	var action string

	if info.ChatUsername != "" {
		action = "resolve"

		args.Set("domain", info.ChatUsername)
	} else {
		action = "privatepost"

		peerID := tg.BotToMTProtoID(info.ChatID)
		args.Set("channel", strconv.FormatInt(peerID, 10))
	}

	args.Set("post", strconv.Itoa(info.PostID))

	return fmt.Sprintf("tg://%s?%s", action, args.Encode())
}

// ProcessChannelPostURIes called on each channel post and should scan for backlinks to bot.
//
// Flow:
//   - resolve chat, if chat is not found finish without error
//   - extract bot /start uries
//   - query files by public link and restrction chat id
//   - update files
func (srv *Chat) ProcessChannelPostURIes(
	ctx context.Context,
	postInfo *ChannelPostInfo,
	uries []string,
) error {
	uries = snip.UniqueizeStrings(uries)

	chat, err := srv.Chat.Query().TelegramID(postInfo.ChatID).One(ctx)
	if errors.Is(err, core.ErrChatNotFound) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "query chat")
	}

	ids, err := ExtractDeepLinkPublicID(srv.Telegram.Self.UserName, uries)
	if err != nil {
		return errors.Wrap(err, "extract deep links payload")
	}

	files, err := srv.File.Query().
		PublicID(ids...).
		RestrictionChatID(chat.ID).
		All(ctx)

	if err != nil {
		return errors.Wrap(err, "query files")
	}

	if err := srv.Txier(ctx, func(ctx context.Context) error {
		link := postInfo.Link()

		for i, file := range files {
			log.Info(ctx, "link post to file", "file_id", file.ID, "post_url", file.LinkedPostURI.String)

			file.LinkedPostURI.SetValid(link)

			if err := srv.File.Update(ctx, file); err != nil {
				return errors.Wrapf(err, "update file #%d", i)
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "update files")
	}

	return nil
}

// ExtractDeepLinkPublicID extract payload from something like{botUsername}?start={payload}
func ExtractDeepLinkPublicID(botUsername string, uries []string) ([]string, error) {
	re := fmt.Sprintf(`%s\?start=([A-Za-z_0-9-]+)`, botUsername)

	linkRegepx, err := regexp.Compile(re)
	if err != nil {
		return nil, errors.Wrap(err, "compile regexp")
	}

	payloads := make([]string, 0, len(uries))
	for _, uri := range uries {
		match := linkRegepx.FindStringSubmatch(uri)
		if match != nil {
			payloads = append(payloads, match[1])
		}
	}

	publicIDs := make([]string, 0, len(payloads))

	for _, payload := range payloads {
		if strings.HasPrefix(payload, "ref_") {
			idx := strings.Index(payload, "-")

			// it's just ref link without file id
			if idx == -1 {
				continue
			}

			// all after -
			payload = payload[idx+1:]
		}

		publicIDs = append(publicIDs, payload)
	}

	return publicIDs, nil
}
