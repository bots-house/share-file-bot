package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserStore struct {
	BaseStore
}

func (store *UserStore) toRow(user *core.User) (*dal.User, error) {
	settings, err := json.Marshal(user.Settings)
	if err != nil {
		return nil, errors.Wrap(err, "marshal settings")
	}

	return &dal.User{
		ID:           int(user.ID),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode,
		IsAdmin:      user.IsAdmin,
		Settings:     settings,
		Ref:          user.Ref,
		JoinedAt:     user.JoinedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (store *UserStore) fromRow(row *dal.User) (*core.User, error) {
	var settings core.UserSettings

	if err := json.Unmarshal(row.Settings, &settings); err != nil {
		return nil, errors.Wrap(err, "ummarshal settings")
	}

	return &core.User{
		ID:           core.UserID(row.ID),
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		Username:     row.Username,
		LanguageCode: row.LanguageCode,
		IsAdmin:      row.IsAdmin,
		Settings:     settings,
		Ref:          row.Ref,
		JoinedAt:     row.JoinedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (store *UserStore) Add(ctx context.Context, user *core.User) error {
	// to row
	row, err := store.toRow(user)
	if err != nil {
		return errors.Wrap(err, "to row")
	}

	// insert
	if err := store.insertOne(ctx, row); err != nil {
		return errors.Wrap(err, "insert query")
	}

	// copy back
	user2, err := store.fromRow(row)
	if err != nil {
		return errors.Wrap(err, "from row")
	}

	*user = *user2

	return nil
}

func (store *UserStore) Find(ctx context.Context, id core.UserID) (*core.User, error) {
	acc, err := dal.FindUser(ctx, store.getExecutor(ctx), int(id))
	if err == sql.ErrNoRows {
		return nil, core.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return store.fromRow(acc)
}

func (store *UserStore) Update(ctx context.Context, user *core.User) error {
	row, err := store.toRow(user)
	if err != nil {
		return errors.Wrap(err, "to row")
	}

	if err := store.updateOne(ctx, row, core.ErrUserNotFound); err != nil {
		return errors.Wrap(err, "update one")
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
	count, err := dal.
		Users(usq.mods...).
		Count(ctx, usq.store.getExecutor(ctx))
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}

func (store *UserStore) RefStats(ctx context.Context) (core.UserRefStats, error) {
	const query = `
		select 
			ref, 
			count(*) as users 
		from 
			"user"
		group by
			ref 
	`

	executor := store.getExecutor(ctx)

	rows, err := executor.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "query rows")
	}
	defer rows.Close()

	result := core.UserRefStats{}

	for rows.Next() {
		item := core.UserRefStatsItem{}

		if err := rows.Scan(
			&item.Ref,
			&item.Count,
		); err != nil {
			return nil, errors.Wrap(err, "scan row")
		}

		result = append(result, item)
	}

	return result, nil
}
