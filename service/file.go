package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bots-house/share-file-bot/core"
	"github.com/friendsofgo/errors"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type File struct {
	File     core.FileStore
	Chat     core.ChatStore
	Telegram *tgbotapi.BotAPI
	Redis    redis.UniversalClient
	Download core.DownloadStore
}

type InputFile struct {
	FileID   string
	Caption  string
	Kind     core.Kind
	MIMEType string
	Name     string
	Size     int

	Metadata core.Metadata
}

type OwnedFile struct {
	*core.File
	Stats *core.FileDownloadStats
}

func (srv *File) newOwnedFile(ctx context.Context, doc *core.File) (*OwnedFile, error) {
	downloadStats, err := srv.Download.GetFileStats(ctx, doc.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get downloads count")
	}

	return &OwnedFile{
		File:  doc,
		Stats: downloadStats,
	}, nil
}

func (srv *File) AddFile(
	ctx context.Context,
	user *core.User,
	in *InputFile,
) (*OwnedFile, error) {
	doc := core.NewFile(
		in.FileID,
		in.Caption,
		in.Kind,
		in.MIMEType,
		in.Size,
		in.Name,
		user.ID,
		user.Settings.LongIDs,
		in.Metadata,
	)

	log.Ctx(ctx).Info().Str("name", in.Name).Int("size", in.Size).Str("kind", in.Kind.String()).Msg("create file")
	if err := srv.File.Add(ctx, doc); err != nil {
		return nil, errors.Wrap(err, "add file to store")
	}

	return srv.newOwnedFile(ctx, doc)
}

type ChatSubRequest struct {
	FileID core.FileID

	Title    string
	Username string
	JoinLink string
}

func (sub *ChatSubRequest) Link() string {
	if sub.Username != "" {
		return "https://t.me/" + sub.Username
	}
	return sub.JoinLink
}

type DownloadResult struct {
	File           *core.File
	OwnedFile      *OwnedFile
	ChatSubRequest *ChatSubRequest
}

var (
	ErrInvalidID = errors.New("invalid file id")
)

func (srv *File) checkFileRestrictionsChat(
	ctx context.Context,
	user *core.User,
	file *core.File,
) (*ChatSubRequest, error) {
	chat, err := srv.Chat.Query().ID(file.Restriction.ChatID).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query chat")
	}

	g, _ := errgroup.WithContext(ctx)

	// query chat
	var tgChat tgbotapi.Chat

	g.Go(func() error {
		var err error
		tgChat, err = srv.Telegram.GetChat(tgbotapi.ChatConfig{
			ChatID: chat.TelegramID,
		})

		if tgChat.UserName == "" {
			link, err := srv.Telegram.GetInviteLink(tgbotapi.ChatConfig{
				ChatID: tgChat.ID,
			})
			if err != nil {
				return errors.Wrap(err, "get chat invite link")
			}

			tgChat.InviteLink = link
		}

		return err
	})

	// query member
	var tgMember tgbotapi.ChatMember

	g.Go(func() error {
		var err error
		tgMember, err = srv.Telegram.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chat.TelegramID,
			UserID: int(user.ID),
		})
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "one of call failed")
	}

	if !(tgMember.IsMember() || tgMember.IsAdministrator() || tgMember.IsCreator()) {
		return &ChatSubRequest{
			FileID:   file.ID,
			Title:    chat.Title,
			Username: tgChat.UserName,
			JoinLink: tgChat.InviteLink,
		}, nil
	}

	return nil, nil
}

func (srv *File) getSubAwaitKey(userID core.UserID, fileID core.FileID) string {
	return fmt.Sprintf("share-file-bot:users:%d:subscription:%d", userID, fileID)
}

func (srv *File) registerSubAwait(ctx context.Context, user *core.User, fileID core.FileID) error {
	key := srv.getSubAwaitKey(user.ID, fileID)
	if err := srv.Redis.Set(ctx, key, true, time.Hour).Err(); err != nil {
		return errors.Wrap(err, "set key")
	}
	return nil
}

func (srv *File) hasSubAwait(ctx context.Context, user *core.User, fileID core.FileID) (bool, error) {
	key := srv.getSubAwaitKey(user.ID, fileID)

	if err := srv.Redis.Get(ctx, key).Err(); err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "get key")
	}

	go func() {
		if err := srv.Redis.Del(ctx, key).Err(); err != nil {
			log.Ctx(ctx).Warn().Str("key", key).Err(err).Msg("can't delete key")
		}
	}()

	return true, nil
}

func (srv *File) toDownloadResult(ctx context.Context, user *core.User, file *core.File) (*DownloadResult, error) {
	// if user is owner of this docs we just display it
	if file.OwnerID == user.ID {
		ownedFile, err := srv.newOwnedFile(ctx, file)
		if err != nil {
			return nil, errors.Wrap(err, "get owned doc")
		}
		return &DownloadResult{
			OwnedFile: ownedFile,
		}, nil
	}

	// check user subscription
	if file.Restriction.HasChatID() {
		sub, err := srv.checkFileRestrictionsChat(ctx, user, file)
		if err != nil {
			return nil, errors.Wrap(err, "check file restrictions chat")
		}

		if sub != nil {
			// add user to subscription await list
			if err := srv.registerSubAwait(ctx, user, file.ID); err != nil {
				return nil, errors.Wrap(err, "can't add user to await list")
			}

			return &DownloadResult{
				ChatSubRequest: sub,
			}, nil
		}
	}

	return srv.RegisterDownload(ctx, user, file)
}

func (srv *File) RegisterDownload(ctx context.Context, user *core.User, file *core.File) (*DownloadResult, error) {
	// register download
	download := core.NewDownload(file.ID, user.ID)

	if file.Restriction.HasChatID() {
		sub, err := srv.hasSubAwait(ctx, user, file.ID)
		if err != nil {
			return nil, errors.Wrap(err, "check sub await")
		}

		download.SetNewSubscription(sub)
	}

	log.Ctx(ctx).Info().Int("file_id", int(file.ID)).Msg("register download")
	if err := srv.Download.Add(ctx, download); err != nil {
		return nil, errors.Wrap(err, "add download to store")
	}

	return &DownloadResult{
		File: file,
	}, nil
}

type ChatRestrictionStatus struct {
	Ok   bool
	Chat *core.Chat
	File *core.File
}

func (srv *File) CheckFileRestrictionsChat(
	ctx context.Context,
	user *core.User,
	id core.FileID,
) (*ChatRestrictionStatus, error) {
	file, err := srv.File.Query().ID(id).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query file by id")
	}

	if !file.Restriction.HasChatID() {
		return &ChatRestrictionStatus{Ok: true, File: file}, nil
	}

	chat, err := srv.Chat.Query().ID(file.Restriction.ChatID).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query chat by restriction")
	}

	member, err := srv.Telegram.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: chat.TelegramID,
		UserID: int(user.ID),
	})

	if err != nil {
		return nil, errors.Wrap(err, "get chat member")
	}

	return &ChatRestrictionStatus{
		Ok:   member.IsMember() || member.IsAdministrator() || member.IsCreator(),
		Chat: chat,
		File: file,
	}, nil
}

func (srv *File) GetFileByID(
	ctx context.Context,
	user *core.User,
	id core.FileID,
) (*DownloadResult, error) {
	doc, err := srv.File.Query().ID(id).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "find file by id")
	}

	return srv.toDownloadResult(ctx, user, doc)
}

func (srv *File) GetFileByPublicID(
	ctx context.Context,
	user *core.User,
	publicID string,
) (*DownloadResult, error) {
	doc, err := srv.File.Query().PublicID(publicID).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "find file by public id")
	}

	return srv.toDownloadResult(ctx, user, doc)
}

func (srv *File) DeleteFile(
	ctx context.Context,
	user *core.User,
	id core.FileID,
) error {
	query := srv.File.Query().
		OwnerID(user.ID).
		ID(id)

	if err := query.Delete(ctx); err != nil {
		return errors.Wrap(err, "delete in store")
	}

	return nil
}

type SetChatRestrictionResult struct {
	Chat    *core.Chat
	File    *core.File
	Disable bool
}

// SetChatRestriction changes chat restriction to specified chat.
func (srv *File) SetChatRestriction(
	ctx context.Context,
	user *core.User,
	fileID core.FileID,
	chatID core.ChatID,
) (*SetChatRestrictionResult, error) {

	log.Ctx(ctx).Info().Int("user_id", int(user.ID)).Int("file_id", int(fileID)).Int("chat_id", int(chatID)).Msg("set chat restriction")
	file, err := srv.File.Query().OwnerID(user.ID).ID(fileID).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query file")
	}

	chat, err := srv.Chat.Query().OwnerID(user.ID).ID(chatID).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query chat")
	}

	disable := file.Restriction.ChatID == chatID

	if disable {
		file.Restriction.ChatID = core.ZeroChatID
	} else {
		file.Restriction.ChatID = chat.ID
	}

	if err := srv.File.Update(ctx, file); err != nil {
		return nil, errors.Wrap(err, "update file")
	}

	return &SetChatRestrictionResult{
		Chat:    chat,
		File:    file,
		Disable: disable,
	}, nil
}
