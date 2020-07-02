package postgres

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DownloadStore struct {
	boil.ContextExecutor
}

func (store *DownloadStore) toRow(dwn *core.Download) *dal.Download {
	return &dal.Download{
		ID:         int(dwn.ID),
		UserID:     null.NewInt(int(dwn.UserID), dwn.UserID != 0),
		DocumentID: null.NewInt(int(dwn.DocumentID), dwn.DocumentID != 0),
		At:         dwn.At,
	}
}

func (store *DownloadStore) fromRow(row *dal.Download) *core.Download {
	return &core.Download{
		ID:         core.DownloadID(row.ID),
		UserID:     core.UserID(row.UserID.Int),
		DocumentID: core.DocumentID(row.DocumentID.Int),
		At:         row.At,
	}
}

func (store *DownloadStore) Add(ctx context.Context, dwn *core.Download) error {
	row := store.toRow(dwn)
	if err := row.Insert(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer()); err != nil {
		return errors.Wrap(err, "insert query")
	}
	*dwn = *store.fromRow(row)
	return nil
}

func (store *DownloadStore) Query() core.DownloadStoreQuery {
	return &downloadStoreQuery{
		store: store,
	}
}

func (store *DownloadStore) GetDownloadStats(ctx context.Context, id core.DocumentID) (*core.DownloadStats, error) {
	const query = `
        select
            count(*) as total, count(distinct user_id) as unique
        from
            download
        where
            document_id = $1
    `

	result := &core.DownloadStats{}

	if err := store.QueryRow(query, id).Scan(
		&result.Total,
		&result.Unique,
	); err != nil {
		return nil, errors.Wrap(err, "count downloads query")
	}

	return result, nil
}

type downloadStoreQuery struct {
	mods  []qm.QueryMod
	store *DownloadStore
}

func (dsq *downloadStoreQuery) DocumentID(id core.DocumentID) core.DownloadStoreQuery {
	dsq.mods = append(dsq.mods, dal.DownloadWhere.DocumentID.EQ(null.IntFrom(int(id))))
	return dsq
}

func (dsq *downloadStoreQuery) Count(ctx context.Context) (int, error) {
	executor := shared.GetExecutorOrDefault(ctx, dsq.store.ContextExecutor)
	count, err := dal.
		Downloads(dsq.mods...).
		Count(ctx, executor)
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
