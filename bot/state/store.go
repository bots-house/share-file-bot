package state

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
)

type Store interface {
	Get(ctx context.Context, id core.UserID) (State, error)
	Set(ctx context.Context, id core.UserID, state State) error
	Del(ctx context.Context, id core.UserID) error
}
