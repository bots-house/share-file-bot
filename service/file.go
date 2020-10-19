package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/pkg/errors"
)

type File struct {
	FileStore     core.FileStore
	DownloadStore core.DownloadStore
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
	downloadStats, err := srv.DownloadStore.GetDownloadStats(ctx, doc.ID)
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
	if err := srv.FileStore.Add(ctx, doc); err != nil {
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
	if err := srv.DownloadStore.Add(ctx, download); err != nil {
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
	doc, err := srv.FileStore.Query().ID(id).One(ctx)
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
	doc, err := srv.FileStore.Query().PublicID(publicID).One(ctx)
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
	query := srv.FileStore.Query().
		OwnerID(user.ID).
		ID(id)

	if err := query.Delete(ctx); err != nil {
		return errors.Wrap(err, "delete in store")
	}

	return nil
}
