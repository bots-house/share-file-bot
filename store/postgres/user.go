package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/bots-house/share-file-bot/store/postgres/shared"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserStore struct {
	boil.ContextExecutor
}

func (store *UserStore) toRow(user *core.User) *dal.User {
	return &dal.User{
		ID:           int(user.ID),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode,
		IsAdmin:      user.IsAdmin,
		JoinedAt:     user.JoinedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func (store *UserStore) fromRow(row *dal.User) *core.User {
	return &core.User{
		ID:           core.UserID(row.ID),
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		Username:     row.Username,
		LanguageCode: row.LanguageCode,
		IsAdmin:      row.IsAdmin,
		JoinedAt:     row.JoinedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (store *UserStore) Add(ctx context.Context, user *core.User) error {
	row := store.toRow(user)
	if err := row.Insert(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer()); err != nil {
		return errors.Wrap(err, "insert query")
	}
	*user = *store.fromRow(row)
	return nil
}

func (store *UserStore) Find(ctx context.Context, id core.UserID) (*core.User, error) {
	acc, err := dal.FindUser(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), int(id))
	if err == sql.ErrNoRows {
		return nil, core.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return store.fromRow(acc), nil
}

func (store *UserStore) Update(ctx context.Context, user *core.User) error {
	row := store.toRow(user)
	n, err := row.Update(ctx, shared.GetExecutorOrDefault(ctx, store.ContextExecutor), boil.Infer())
	if err != nil {
		return errors.Wrap(err, "update query")
	}
	if n == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

func (store *UserStore) Query() core.UserStoreQuery {
	return &userStoreQuery{store: store}
}

type userStoreQuery struct {
	mods  []qm.QueryMod
	store *UserStore
}

func (usq *userStoreQuery) Count(ctx context.Context) (int, error) {
	executor := shared.GetExecutorOrDefault(ctx, usq.store.ContextExecutor)
	count, err := dal.
		Users(usq.mods...).
		Count(ctx, executor)
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
