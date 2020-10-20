package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/bots-house/share-file-bot/core"
	"github.com/stretchr/testify/require"
)

func newFakeUserInStore(t *testing.T, pg *Postgres) *core.User {
	t.Helper()

	user := newFakeUser()

	err := pg.User().Add(context.Background(), user)

	if err != nil {
		t.Fatalf("can't add fake user to store: %v", err)
	}

	return user
}

func TestChatStore_Add(t *testing.T) {
	ctx, pg := newPostgres(t)
	store := pg.Chat()

	user := newFakeUserInStore(t, pg)

	chat := core.NewChat(
		-100123456,
		"test",
		core.ChatTypeChannel,
		user.ID,
	)

	err := store.Add(ctx, chat)

	require.NoError(t, err)
	require.NotZero(t, chat.ID, "chat id should be not zero")
	require.False(t, chat.UpdatedAt.Valid, "updated at should be null")
}

func TestChatStore_Update(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		ctx, pg := newPostgres(t)
		store := pg.Chat()

		user := newFakeUserInStore(t, pg)

		chat := core.NewChat(
			-100123456,
			"test",
			core.ChatTypeChannel,
			user.ID,
		)

		// create & update chat
		{
			err := store.Add(ctx, chat)
			require.NoError(t, err)
			require.NotZero(t, chat.ID, "chat id should be not zero")

			updated := chat.Patch(func(chat *core.Chat) {
				chat.Title = "test 2"
			})

			require.True(t, updated, "should be updated")
			require.True(t, chat.UpdatedAt.Valid, "updated at should be not null")
		}

		err := store.Update(ctx, chat)
		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx, pg := newPostgres(t)
		store := pg.Chat()

		user := newFakeUserInStore(t, pg)

		chat := core.NewChat(
			-100123456,
			"test",
			core.ChatTypeChannel,
			user.ID,
		)

		chat.ID = 1111

		err := store.Update(ctx, chat)
		require.True(t, errors.Is(err, core.ErrChatNotFound), "error should be core.ErrChatNotFound")
	})
}
