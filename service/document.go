package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/pkg/secretid"
	"github.com/pkg/errors"
)

type Document struct {
	SecretID      secretid.SecretID
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
	SecretID string
	Stats    *core.DownloadStats
}

func (srv *Document) newOwnedDocument(ctx context.Context, doc *core.Document) (*OwnedDocument, error) {
	secretID := srv.SecretID.Encode(int(doc.ID))

	downloadStats, err := srv.DownloadStore.GetDownloadStats(ctx, doc.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get downloads count")
	}

	return &OwnedDocument{
		Document: doc,
		SecretID: secretID,
		Stats:    downloadStats,
	}, nil
}

func (srv *Document) AddDocument(
	ctx context.Context,
	user *core.User,
	in *InputDocument,
) (*OwnedDocument, error) {
	doc := core.NewDocument(in.FileID, in.Caption, in.MIMEType, in.Size, in.Name, user.ID)

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

func (srv *Document) DownloadDocument(
	ctx context.Context,
	user *core.User,
	hash string,
) (*DownloadResult, error) {
	id, err := srv.SecretID.Decode(hash)
	if err != nil {
		log.Warn(ctx, "invalid document id", "err", err)
		return nil, ErrInvalidID
	}

	docID := core.DocumentID(id)

	doc, err := srv.DocumentStore.Find(ctx, docID)
	if err != nil {
		return nil, errors.Wrap(err, "find document")
	}

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
	if err := srv.DownloadStore.Add(ctx, download); err != nil {
		return nil, errors.Wrap(err, "download result")
	}

	return &DownloadResult{
		Document: doc,
	}, nil
}
