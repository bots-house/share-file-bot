package state

import (
	"context"
	"strconv"
	"strings"

	"github.com/bots-house/share-file-bot/core"
	"github.com/friendsofgo/errors"
	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	prefix string
	client redis.UniversalClient
}

func NewRedisStore(client redis.UniversalClient, prefix string) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

func (rs *RedisStore) getKey(id core.UserID) string {
	parts := []string{
		rs.prefix,
		"users",
		strconv.Itoa(int(id)),
		"state",
	}

	return strings.Join(parts, ":")
}

func (rs *RedisStore) Set(ctx context.Context, id core.UserID, state State) error {
	key := rs.getKey(id)
	val := strconv.Itoa(int(state))

	if err := rs.client.Set(ctx, key, val, 0).Err(); err != nil {
		return errors.Wrap(err, "set key failed")
	}
	return nil
}

func (rs *RedisStore) Get(ctx context.Context, id core.UserID) (State, error) {
	val, err := rs.client.Get(ctx, rs.getKey(id)).Result()
	if err == redis.Nil {
		return Empty, nil
	} else if err != nil {
		return Empty, errors.Wrap(err, "get value")
	}

	state, err := strconv.Atoi(val)
	if err != nil {
		return Empty, errors.Wrap(err, "parse value")
	}

	return State(state), nil
}

func (rs *RedisStore) Del(ctx context.Context, id core.UserID) error {
	key := rs.getKey(id)

	if err := rs.client.Del(ctx, key).Err(); err == redis.Nil {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "del key")
	}

	return nil
}
