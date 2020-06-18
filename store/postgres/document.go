package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
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
		Size:      doc.Size,
		Name:      doc.Name,
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
		Size:      row.Size,
		Name:      row.Name,
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

func (store *DocumentStore) Query() core.DocumentStoreQuery {
	return &documentStoreQuery{store: store}
}

type documentStoreQuery struct {
	mods  []qm.QueryMod
	store *DocumentStore
}

func (dsq *documentStoreQuery) ID(id core.DocumentID) core.DocumentStoreQuery {
	dsq.mods = append(dsq.mods, dal.DocumentWhere.ID.EQ(int(id)))
	return dsq
}

func (dsq *documentStoreQuery) OwnerID(id core.UserID) core.DocumentStoreQuery {
	dsq.mods = append(dsq.mods, dal.DocumentWhere.OwnerID.EQ(int(id)))
	return dsq
}

func (dsq *documentStoreQuery) Delete(ctx context.Context) error {
	executor := shared.GetExecutorOrDefault(ctx, dsq.store.ContextExecutor)
	count, err := dal.
		Documents(dsq.mods...).
		DeleteAll(ctx, executor)
	if err != nil {
		return errors.Wrap(err, "delete query")
	}
	if count == 0 {
		return core.ErrDocumentNotFound
	}
	return nil
}

func (dsq *documentStoreQuery) Count(ctx context.Context) (int, error) {
	executor := shared.GetExecutorOrDefault(ctx, dsq.store.ContextExecutor)
	count, err := dal.
		Documents(dsq.mods...).
		Count(ctx, executor)
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
