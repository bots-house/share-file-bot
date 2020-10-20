package postgres

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/bots-house/share-file-bot/store"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
)

// BaseStore define base store
type BaseStore struct {
	*sql.DB
	store.Txier
}

func (bs *BaseStore) getExecutor(ctx context.Context) boil.ContextExecutor {
	return shared.GetExecutorOrDefault(ctx, bs)
}

// type deletableRow interface {
// 	Delete(
// 		ctx context.Context,
// 		exec boil.ContextExecutor,
// 	) (int64, error)
// }

// func (bs *BaseStore) deleteOne(
// 	ctx context.Context,
// 	row deletableRow,
// 	notFoundErr error,
// ) error {
// 	return bs.Txier(ctx, func(ctx context.Context) error {
// 		count, err := row.Delete(
// 			ctx,
// 			bs.getExecutor(ctx),
// 		)

// 		if err != nil {
// 			return errors.Wrap(err, "exec")
// 		}

// 		switch {
// 		case count == 0:
// 			return notFoundErr
// 		case count > 1:
// 			return store.ErrTooManyAffectedRows
// 		}

// 		return nil
// 	})
// }

type updatetableRow interface {
	Update(
		ctx context.Context,
		exec boil.ContextExecutor,
		columns boil.Columns,
	) (int64, error)
}

type insertableRow interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error
}

func (bs *BaseStore) insertOne(
	ctx context.Context,
	row insertableRow,
) error {
	return row.Insert(ctx, bs.getExecutor(ctx), boil.Infer())
}

func (bs *BaseStore) updateOne(
	ctx context.Context,
	row updatetableRow,
	notFoundErr error,
) error {
	return bs.Txier(ctx, func(ctx context.Context) error {
		count, err := row.Update(
			ctx,
			bs.getExecutor(ctx),
			boil.Infer(),
		)

		if err != nil {
			return errors.Wrap(err, "exec")
		}

		switch {
		case count == 0:
			return notFoundErr
		case count > 1:
			return store.ErrTooManyAffectedRows
		}

		return nil
	})
}
