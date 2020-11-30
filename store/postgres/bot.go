package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BotStore struct {
	BaseStore
}

func (store *BotStore) toRow(bot *core.Bot) *dal.Bot {
	return &dal.Bot{
		ID:       int(bot.ID),
		Username: bot.Username,
		Token:    bot.Token,
		OwnerID:  int(bot.OwnerID),
		LinkedAt: bot.LinkedAt,
	}
}

func (store *BotStore) fromRow(row *dal.Bot) *core.Bot {
	return &core.Bot{
		ID:       core.BotID(row.ID),
		Username: row.Username,
		Token:    row.Token,
		OwnerID:  core.UserID(row.OwnerID),
		LinkedAt: row.LinkedAt,
	}
}

func (store *BotStore) fromRows(rows []*dal.Bot) []*core.Bot {
	result := make([]*core.Bot, len(rows))

	for i, row := range rows {
		result[i] = store.fromRow(row)
	}

	return result
}

func (store *BotStore) Add(ctx context.Context, bot *core.Bot) error {
	row := store.toRow(bot)

	if err := store.insertOne(ctx, row); err != nil {
		return errors.Wrap(err, "insert query")
	}

	// copy back
	bot2 := store.fromRow(row)

	*bot = *bot2

	return nil
}

func (store *BotStore) Update(ctx context.Context, bot *core.Bot) error {
	row := store.toRow(bot)

	if err := store.updateOne(ctx, row, core.ErrBotNotFound); err != nil {
		return errors.Wrap(err, "update one")
	}

	return nil
}

func (store *BotStore) Query() core.BotStoreQuery {
	return &botStoreQuery{store: store}
}

type botStoreQuery struct {
	mods  []qm.QueryMod
	store *BotStore
}

func (bsq *botStoreQuery) ID(id core.BotID) core.BotStoreQuery {
	bsq.mods = append(bsq.mods, dal.BotWhere.ID.EQ(int(id)))
	return bsq
}

func (bsq *botStoreQuery) OwnerID(id core.UserID) core.BotStoreQuery {
	bsq.mods = append(bsq.mods, dal.BotWhere.OwnerID.EQ(int(id)))
	return bsq
}

func (bsq *botStoreQuery) One(ctx context.Context) (*core.Bot, error) {
	executor := bsq.store.getExecutor(ctx)

	bot, err := dal.Bots(bsq.mods...).One(ctx, executor)
	if err == sql.ErrNoRows {
		return nil, core.ErrBotNotFound
	} else if err != nil {
		return nil, err
	}

	return bsq.store.fromRow(bot), nil
}

func (bsq *botStoreQuery) All(ctx context.Context) ([]*core.Bot, error) {
	executor := bsq.store.getExecutor(ctx)

	rows, err := dal.Bots(bsq.mods...).All(ctx, executor)
	if err != nil {
		return nil, err
	}

	return bsq.store.fromRows(rows), nil
}

func (bsq *botStoreQuery) Delete(ctx context.Context) error {
	executor := bsq.store.getExecutor(ctx)
	count, err := dal.
		Bots(bsq.mods...).
		DeleteAll(ctx, executor)

	if err != nil {
		return errors.Wrap(err, "delete query")
	}

	if count == 0 {
		return core.ErrBotNotFound
	}

	return nil
}

func (bsq *botStoreQuery) Count(ctx context.Context) (int, error) {
	count, err := dal.
		Bots(bsq.mods...).
		Count(ctx, bsq.store.getExecutor(ctx))
	if err != nil {
		return 0, errors.Wrap(err, "count query")
	}

	return int(count), nil
}
