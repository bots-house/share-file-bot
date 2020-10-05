package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type FileStore struct {
	boil.ContextExecutor
}

func (store *FileStore) toRow(file *core.File) *dal.File {
	return &dal.File{
		ID:        int(file.ID),
		FileID:    file.TelegramID,
		PublicID:  file.PublicID,
		Caption:   file.Caption,
		MimeType:  file.MIMEType,
		Size:      file.Size,
		Name:      file.Name,
		OwnerID:   int(file.OwnerID),
		CreatedAt: file.CreatedAt,
	}
}

func (store *FileStore) fromRow(row *dal.File) *core.File {
	return &core.File{
		ID:         core.FileID(row.ID),
		TelegramID: row.FileID,
		PublicID:   row.PublicID,
		Caption:    row.Caption,
		MIMEType:   row.MimeType,
		Size:       row.Size,
		Name:       row.Name,
		OwnerID:    core.UserID(row.OwnerID),
		CreatedAt:  row.CreatedAt,
	}
}

func (store *FileStore) Add(ctx context.Context, doc *core.File) error {
	for {
		if err := store.add(ctx, doc); err != nil {
			if isFilePublicIDCollisionErr(err) {
				currID := doc.PublicID
				doc.RegenPublicID()
				log.Warn(ctx, "collison when insert doc", "curr_id", currID, "next_id", doc.PublicID)
				continue
			} else {
				return err
			}
		}

		return nil
	}
}

func (store *FileStore) add(ctx context.Context, file *core.File) error {
	row := store.toRow(file)
	if err := row.Insert(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer()); err != nil {
		return errors.Wrap(err, "insert query")
	}
	*file = *store.fromRow(row)
	return nil
}

func (store *FileStore) Update(ctx context.Context, file *core.File) error {
	row := store.toRow(file)
	n, err := row.Update(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer())
	if err != nil {
		return errors.Wrap(err, "update query")
	}
	if n == 0 {
		return core.ErrFileNotFound
	}
	return nil
}

func (store *FileStore) Query() core.FileStoreQuery {
	return &fileStoreQuery{store: store}
}

type fileStoreQuery struct {
	mods  []qm.QueryMod
	store *FileStore
}

func (fsq *fileStoreQuery) ID(id core.FileID) core.FileStoreQuery {
	fsq.mods = append(fsq.mods, dal.FileWhere.ID.EQ(int(id)))
	return fsq
}

func (fsq *fileStoreQuery) PublicID(id string) core.FileStoreQuery {
	fsq.mods = append(fsq.mods, dal.FileWhere.PublicID.EQ(id))
	return fsq
}

func (fsq *fileStoreQuery) OwnerID(id core.UserID) core.FileStoreQuery {
	fsq.mods = append(fsq.mods, dal.FileWhere.OwnerID.EQ(int(id)))
	return fsq
}

func (fsq *fileStoreQuery) One(ctx context.Context) (*core.File, error) {
	executor := shared.GetExecutorOrDefault(ctx, fsq.store.ContextExecutor)

	doc, err := dal.Files(fsq.mods...).One(ctx, executor)
	if err == sql.ErrNoRows {
		return nil, core.ErrFileNotFound
	} else if err != nil {
		return nil, err
	}

	return fsq.store.fromRow(doc), nil
}

func (fsq *fileStoreQuery) Delete(ctx context.Context) error {
	executor := shared.GetExecutorOrDefault(ctx, fsq.store.ContextExecutor)
	count, err := dal.
		Files(fsq.mods...).
		DeleteAll(ctx, executor)
	if err != nil {
		return errors.Wrap(err, "delete query")
	}
	if count == 0 {
		return core.ErrFileNotFound
	}
	return nil
}

func (fsq *fileStoreQuery) Count(ctx context.Context) (int, error) {
	executor := shared.GetExecutorOrDefault(ctx, fsq.store.ContextExecutor)
	count, err := dal.
		Files(fsq.mods...).
		Count(ctx, executor)
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
