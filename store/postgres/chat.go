package postgres

import (
	"context"
	"database/sql"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/service"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ChatStore struct {
	BaseStore
}

func (store *ChatStore) toRow(chat *core.Chat) *dal.Chat {
	return &dal.Chat{
		ID:         int(chat.ID),
		TelegramID: chat.TelegramID,
		Title:      chat.Title,
		Type:       chat.Type.String(),
		OwnerID:    int(chat.OwnerID),
		LinkedAt:   chat.LinkedAt,
		UpdatedAt:  chat.UpdatedAt,
	}
}

func (store *ChatStore) fromRow(row *dal.Chat) (*core.Chat, error) {
	chatType, err := core.ParseChatType(row.Type)
	if err != nil {
		return nil, errors.Wrap(err, "parse chat type")
	}

	return &core.Chat{
		ID:         core.ChatID(row.ID),
		TelegramID: row.TelegramID,
		Title:      row.Title,
		Type:       chatType,
		OwnerID:    core.UserID(row.OwnerID),
		LinkedAt:   row.LinkedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (store *ChatStore) fromRowSlice(rows dal.ChatSlice) ([]*core.Chat, error) {
	result := make([]*core.Chat, len(rows))

	for i, row := range rows {
		chat, err := store.fromRow(row)
		if err != nil {
			return nil, errors.Wrapf(err, "from row #%d", i)
		}
		result[i] = chat
	}

	return result, nil
}

// Add chat to store.
func (store *ChatStore) Add(ctx context.Context, chat *core.Chat) error {
	row := store.toRow(chat)
	if err := store.insertOne(ctx, row); err != nil {
		if isChatAlreadyConnectedError(err) {
			return service.ErrChatAlreadyConnected
		}
		return errors.Wrap(err, "insert query")
	}

	newChat, err := store.fromRow(row)
	if err != nil {
		return errors.Wrap(err, "from row")
	}

	*chat = *newChat

	return nil
}

// Update chat in store.
func (store *ChatStore) Update(ctx context.Context, chat *core.Chat) error {
	row := store.toRow(chat)

	if err := store.updateOne(ctx, row, core.ErrChatNotFound); err != nil {
		return errors.Wrap(err, "update one")
	}

	return nil
}

func (store *ChatStore) Query() core.ChatStoreQuery {
	return &ChatStoreQuery{
		Store: store,
	}
}

type ChatStoreQuery struct {
	Store *ChatStore
	Mods  []qm.QueryMod
}

// ID Filter
func (csq *ChatStoreQuery) ID(ids ...core.ChatID) core.ChatStoreQuery {
	idsInt := make([]int, len(ids))
	for i, v := range ids {
		idsInt[i] = int(v)
	}

	csq.Mods = append(csq.Mods, dal.ChatWhere.ID.IN(idsInt))

	return csq
}

// TelegramID filter
func (csq *ChatStoreQuery) TelegramID(id int64) core.ChatStoreQuery {
	csq.Mods = append(csq.Mods, dal.ChatWhere.TelegramID.EQ(id))
	return csq
}

// OwnerID filter
func (csq *ChatStoreQuery) OwnerID(id core.UserID) core.ChatStoreQuery {
	csq.Mods = append(csq.Mods, dal.ChatWhere.OwnerID.EQ(int(id)))
	return csq
}

func (csq *ChatStoreQuery) Delete(ctx context.Context) (int, error) {
	count, err := dal.Chats(csq.Mods...).
		DeleteAll(ctx, csq.Store.getExecutor(ctx))

	return int(count), err
}

// One return only one item from store.
func (csq *ChatStoreQuery) One(ctx context.Context) (*core.Chat, error) {
	row, err := dal.Chats(csq.Mods...).One(ctx, csq.Store.getExecutor(ctx))
	if err == sql.ErrNoRows {
		return nil, core.ErrChatNotFound
	} else if err != nil {
		return nil, err
	}

	return csq.Store.fromRow(row)
}

// All query items from store.
func (csq *ChatStoreQuery) All(ctx context.Context) ([]*core.Chat, error) {
	rows, err := dal.Chats(csq.Mods...).All(ctx, csq.Store.getExecutor(ctx))
	if err != nil {
		return nil, err
	}
	return csq.Store.fromRowSlice(rows)
}

// Count items in store.
func (csq *ChatStoreQuery) Count(ctx context.Context) (int, error) {
	count, err := dal.Chats(csq.Mods...).Count(ctx, csq.Store.getExecutor(ctx))
	if err != nil {
		return -1, err
	}

	return int(count), nil
}
