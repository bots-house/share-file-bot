package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/pkg/errors"
)

type Document struct {
	DocumentStore core.DocumentStore
	DownloadStore core.DownloadStore
}

type InputDocument struct {
	FileID   string
	Caption  string
	MIMEType string
	Name     string
	Size     int
}

type OwnedDocument struct {
	*core.Document
	Stats *core.DownloadStats
}

func (srv *Document) newOwnedDocument(ctx context.Context, doc *core.Document) (*OwnedDocument, error) {
	downloadStats, err := srv.DownloadStore.GetDownloadStats(ctx, doc.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get downloads count")
	}

	return &OwnedDocument{
		Document: doc,
		Stats:    downloadStats,
	}, nil
}

func (srv *Document) AddDocument(
	ctx context.Context,
	user *core.User,
	in *InputDocument,
) (*OwnedDocument, error) {
	doc := core.NewDocument(
		in.FileID,
		in.Caption,
		in.MIMEType,
		in.Size,
		in.Name,
		user.ID,
		user.Settings.LongIDs,
	)

	log.Info(ctx, "create document", "name", in.Name, "size", in.Size)
	if err := srv.DocumentStore.Add(ctx, doc); err != nil {
		return nil, errors.Wrap(err, "add document to store")
	}

	return srv.newOwnedDocument(ctx, doc)
}

type DownloadResult struct {
	Document      *core.Document
	OwnedDocument *OwnedDocument
}

var (
	ErrInvalidID = errors.New("invalid document id")
)

func (srv *Document) toDownloadResult(ctx context.Context, user *core.User, doc *core.Document) (*DownloadResult, error) {
	// if user is owner of this docs we just display it
	if doc.OwnerID == user.ID {
		ownedDoc, err := srv.newOwnedDocument(ctx, doc)
		if err != nil {
			return nil, errors.Wrap(err, "get owned doc")
		}
		return &DownloadResult{
			OwnedDocument: ownedDoc,
		}, nil
	}

	// register download
	download := core.NewDownload(doc.ID, user.ID)

	log.Info(ctx, "register download", "document_id", doc.ID)
	if err := srv.DownloadStore.Add(ctx, download); err != nil {
		return nil, errors.Wrap(err, "download result")
	}

	return &DownloadResult{
		Document: doc,
	}, nil
}

func (srv *Document) GetDocumentByID(
	ctx context.Context,
	user *core.User,
	id core.DocumentID,
) (*DownloadResult, error) {
	doc, err := srv.DocumentStore.Query().ID(id).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "find document by id")
	}

	return srv.toDownloadResult(ctx, user, doc)
}

func (srv *Document) GetDocumentByPublicID(
	ctx context.Context,
	user *core.User,
	publicID string,
) (*DownloadResult, error) {
	doc, err := srv.DocumentStore.Query().PublicID(publicID).One(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "find document by public id")
	}

	return srv.toDownloadResult(ctx, user, doc)
}

func (srv *Document) DeleteDocument(
	ctx context.Context,
	user *core.User,
	id core.DocumentID,
) error {
	query := srv.DocumentStore.Query().
		OwnerID(user.ID).
		ID(id)

	if err := query.Delete(ctx); err != nil {
		return errors.Wrap(err, "delete in store")
	}

	return nil
}
