package postgres

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DownloadStore struct {
	BaseStore
}

func (store *DownloadStore) toRow(dwn *core.Download) *dal.Download {
	return &dal.Download{
		ID:              int(dwn.ID),
		UserID:          null.NewInt(int(dwn.UserID), dwn.UserID != 0),
		FileID:          null.NewInt(int(dwn.FileID), dwn.FileID != 0),
		NewSubscription: dwn.NewSubscription,
		At:              dwn.At,
	}
}

func (store *DownloadStore) fromRow(row *dal.Download) *core.Download {
	return &core.Download{
		ID:              core.DownloadID(row.ID),
		UserID:          core.UserID(row.UserID.Int),
		FileID:          core.FileID(row.FileID.Int),
		NewSubscription: row.NewSubscription,
		At:              row.At,
	}
}

func (store *DownloadStore) Add(ctx context.Context, dwn *core.Download) error {
	row := store.toRow(dwn)
	if err := store.insertOne(ctx, row); err != nil {
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

func (store *DownloadStore) GetFileDownloadStats(ctx context.Context, id core.FileID) (*core.DownloadStats, error) {
	const query = `
	select
		count(user_id) as total, 
		count(distinct user_id) as unique, 
		coalesce(sum(case when new_subscription = true then 1 else 0 end), 0) as new_subscription, 
		coalesce(sum(case when new_subscription = false then 1 else 0 end), 0) as with_subscription
	from
		download
	where
		file_id = $1
    `

	result := &core.DownloadStats{}

	executor := store.getExecutor(ctx)

	if err := executor.QueryRowContext(ctx, query, id).Scan(
		&result.Total,
		&result.Unique,
		&result.NewSubscription,
		&result.WithSubscription,
	); err != nil {
		return nil, errors.Wrap(err, "count downloads query")
	}

	return result, nil
}

type downloadStoreQuery struct {
	mods  []qm.QueryMod
	store *DownloadStore
}

func (dsq *downloadStoreQuery) FileID(id core.FileID) core.DownloadStoreQuery {
	dsq.mods = append(dsq.mods, dal.DownloadWhere.FileID.EQ(null.IntFrom(int(id))))
	return dsq
}

func (dsq *downloadStoreQuery) Count(ctx context.Context) (int, error) {
	count, err := dal.
		Downloads(dsq.mods...).
		Count(ctx, dsq.store.getExecutor(ctx))
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
