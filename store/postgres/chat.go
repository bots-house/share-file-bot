package postgres

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/bots-house/share-file-bot/store/postgres/dal"
	"github.com/friendsofgo/errors"
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

// Add chat to store.
func (store *ChatStore) Add(ctx context.Context, chat *core.Chat) error {
	row := store.toRow(chat)
	if err := store.insertOne(ctx, row); err != nil {
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
