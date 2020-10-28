package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/pkg/log"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type FileStore struct {
	BaseStore
}

func (store *FileStore) toRow(file *core.File) (*dal.File, error) {
	metadata, err := json.Marshal(file.Metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal metadata")
	}

	return &dal.File{
		ID:                 int(file.ID),
		FileID:             file.TelegramID,
		PublicID:           file.PublicID,
		Caption:            file.Caption,
		MimeType:           file.MIMEType,
		Kind:               file.Kind.String(),
		RestrictionsChatID: null.NewInt(int(file.Restriction.ChatID), file.Restriction.ChatID != 0),
		Metadata:           metadata,
		Size:               file.Size,
		Name:               file.Name,
		OwnerID:            int(file.OwnerID),
		CreatedAt:          file.CreatedAt,
	}, nil
}

func (store *FileStore) fromRow(row *dal.File) (*core.File, error) {
	kind, err := core.ParseKind(row.Kind)
	if err != nil {
		return nil, err
	}

	var metadata core.Metadata

	if err := row.Metadata.Unmarshal(&metadata); err != nil {
		return nil, errors.Wrap(err, "unmarshal metadata")
	}

	return &core.File{
		ID:         core.FileID(row.ID),
		TelegramID: row.FileID,
		PublicID:   row.PublicID,
		Caption:    row.Caption,
		Kind:       kind,
		Metadata:   metadata,
		MIMEType:   row.MimeType,
		Restriction: core.DownloadRestrictions{
			ChatID: core.ChatID(row.RestrictionsChatID.Int),
		},
		Size:      row.Size,
		Name:      row.Name,
		OwnerID:   core.UserID(row.OwnerID),
		CreatedAt: row.CreatedAt,
	}, nil
}

func (store *FileStore) Add(ctx context.Context, file *core.File) error {
	for {
		if err := store.add(ctx, file); err != nil {
			if isFilePublicIDCollisionErr(err) {
				currID := file.PublicID
				file.RegenPublicID()
				log.Warn(ctx, "collison when insert doc", "curr_id", currID, "next_id", file.PublicID)
				continue
			} else {
				return err
			}
		}

		return nil
	}
}

func (store *FileStore) add(ctx context.Context, file *core.File) error {
	row, err := store.toRow(file)
	if err != nil {
		return errors.Wrap(err, "to row")
	}

	if err := store.insertOne(ctx, row); err != nil {
		return errors.Wrap(err, "insert query")
	}

	newFile, err := store.fromRow(row)
	if err != nil {
		return errors.Wrap(err, "from row")
	}

	*file = *newFile

	return nil
}

func (store *FileStore) Update(ctx context.Context, file *core.File) error {
	row, err := store.toRow(file)
	if err != nil {
		return errors.Wrap(err, "to row")
	}

	if err := store.updateOne(ctx, row, core.ErrFileNotFound); err != nil {
		return errors.Wrap(err, "update one")
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

func (fsq *fileStoreQuery) RestrictionChatID(id core.ChatID) core.FileStoreQuery {
	v := null.NewInt(int(id), id != core.ZeroChatID)
	fsq.mods = append(fsq.mods, dal.FileWhere.RestrictionsChatID.EQ(v))
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
	executor := fsq.store.getExecutor(ctx)

	file, err := dal.Files(fsq.mods...).One(ctx, executor)
	if err == sql.ErrNoRows {
		return nil, core.ErrFileNotFound
	} else if err != nil {
		return nil, err
	}

	return fsq.store.fromRow(file)
}

func (fsq *fileStoreQuery) Delete(ctx context.Context) error {
	executor := fsq.store.getExecutor(ctx)
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
	executor := fsq.store.getExecutor(ctx)

	count, err := dal.
		Files(fsq.mods...).
		Count(ctx, executor)

	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
