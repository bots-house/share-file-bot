package bot

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
)

type contextKey int

const (
    userCtxKey contextKey = iota
)

func withUser(ctx context.Context, user *core.User) context.Context {
    return context.WithValue(ctx, userCtxKey, user)
}

func getUserCtx(ctx context.Context) *core.User {
    return ctx.Value(userCtxKey).(*core.User)
}
