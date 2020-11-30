package core

import (
	"context"
	"errors"
	"time"
)

// BotID unique bot id, equal Telegram ID.
type BotID int

// Bot define type of bot.
type Bot struct {
	// Unique bot id
	ID BotID

	// Username of bot
	Username string

	// Token of bot
	Token string

	// Owner of the bot
	OwnerID UserID

	// Time when bot was linked to Share File Bot
	LinkedAt time.Time
}

var (
	// ErrBotNotFound returned by store when bot is not found
	ErrBotNotFound = errors.New("bot not found")
)

// BotStore define persistance for bots.
type BotStore interface {
	// Add bot to store.
	Add(ctx context.Context, bot *Bot) error

	// Update bot in store.
	Update(ctx context.Context, bot *Bot) error

	// Query bots
	Query() BotStoreQuery
}

// BotStoreQuery define interface for custom query.
type BotStoreQuery interface {
	// Filter bots by ID
	ID(id BotID) BotStoreQuery

	// Filter bots by Owner ID
	OwnerID(userID UserID) BotStoreQuery

	// One bot
	One(ctx context.Context) (*Bot, error)

	// All bots matching query
	All(ctx context.Context) ([]*Bot, error)

	// Delete bots matching query
	Delete(ctx context.Context) error

	// Count bots
	Count(ctx context.Context) (int, error)
}
