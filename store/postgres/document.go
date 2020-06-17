package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

type DocumentStore struct {
	boil.ContextExecutor
}

func (store *DocumentStore) toRow(doc *core.Document) *dal.Document {
	return &dal.Document{
		ID:        int(doc.ID),
		FileID:    doc.FileID,
		Caption:   doc.Caption,
		MimeType:  doc.MIMEType,
		OwnerID:   int(doc.OwnerID),
		CreatedAt: doc.CreatedAt,
	}
}

func (store *DocumentStore) fromRow(row *dal.Document) *core.Document {
	return &core.Document{
		ID:        core.DocumentID(row.ID),
		FileID:    row.FileID,
		Caption:   row.Caption,
		MIMEType:  row.MimeType,
		OwnerID:   core.UserID(row.OwnerID),
		CreatedAt: row.CreatedAt,
	}
}

func (store *DocumentStore) Add(ctx context.Context, doc *core.Document) error {
	row := store.toRow(doc)
	if err := row.Insert(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer()); err != nil {
		return errors.Wrap(err, "insert query")
	}
	*doc = *store.fromRow(row)
	return nil
}

func (store *DocumentStore) Find(ctx context.Context, id core.DocumentID) (*core.Document, error) {
	doc, err := dal.FindDocument(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), int(id))
	if err == sql.ErrNoRows {
		return nil, core.ErrDocumentNotFound
	} else if err != nil {
		return nil, err
	}

	return store.fromRow(doc), nil
}

func (store *DocumentStore) Update(ctx context.Context, doc *core.Document) error {
	row := store.toRow(doc)
	n, err := row.Update(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer())
	if err != nil {
		return errors.Wrap(err, "update query")
	}
	if n == 0 {
		return core.ErrDocumentNotFound
	}
	return nil
}
