package postgres

import (
	"testing"

	"github.com/bots-house/share-file-bot/core"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func newFakeUser() *core.User {
	return &core.User{
		ID:           core.UserID(gofakeit.Number(1, 9999999)),
		FirstName:    gofakeit.FirstName(),
		LastName:     null.StringFrom(gofakeit.LastName()),
		Username:     null.StringFrom(gofakeit.Username()),
		LanguageCode: gofakeit.Language(),
		JoinedAt:     gofakeit.Date(),
	}
}

func TestUserStore_Add(t *testing.T) {
	ctx, pg := newPostgres(t)
	store := pg.User()

	user := newFakeUser()

	err := store.Add(ctx, user)
	assert.NoError(t, err, "add should finish without error")
}

func TestUserStore_Find(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		ctx, pg := newPostgres(t)
		store := pg.User()

		// insert user
		user := newFakeUser()
		err := store.Add(ctx, user)
		require.NoError(t, err, "add should finish without error")

		// find user
		user2, err := store.Find(ctx, user.ID)
		require.NoError(t, err, "find user should finish without error")

		assert.Equal(t, user.ID, user2.ID)
		assert.Equal(t, user.FirstName, user2.FirstName)
		assert.Equal(t, user.LastName, user2.LastName)
		assert.Equal(t, user.Username, user2.Username)
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx, pg := newPostgres(t)
		store := pg.User()

		// find user
		user, err := store.Find(ctx, core.UserID(12345))

		assert.Equal(t, core.ErrUserNotFound, err, "error should be core.ErrUserNotFound")
		assert.Nil(t, user)
	})
}
