package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/store"
	"github.com/friendsofgo/errors"
)

type File struct {
	File     core.FileStore
	Chat     core.ChatStore
	Txier    store.Txier
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
	Stats *core.DownloadStats
}

func (srv *File) newOwnedFile(ctx context.Context, doc *core.File) (*OwnedFile, error) {
	downloadStats, err := srv.Download.GetDownloadStats(ctx, doc.ID)
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

	log.Info(ctx, "create file",
		"name", in.Name,
		"size", in.Size,
		"kind", in.Kind.String(),
	)
	if err := srv.File.Add(ctx, doc); err != nil {
		return nil, errors.Wrap(err, "add file to store")
	}

	return srv.newOwnedFile(ctx, doc)
}

type DownloadResult struct {
	File      *core.File
	OwnedFile *OwnedFile
}

var (
	ErrInvalidID = errors.New("invalid file id")
)

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

	// register download
	download := core.NewDownload(file.ID, user.ID)

	log.Info(ctx, "register download", "file_id", file.ID)
	if err := srv.Download.Add(ctx, download); err != nil {
		return nil, errors.Wrap(err, "download result")
	}

	return &DownloadResult{
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

	log.Info(ctx,
		"set chat restriction",
		"user_id", user.ID,
		"file_id", fileID,
		"chat_id", chatID,
	)

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
