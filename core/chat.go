package core

import (
	"context"
	"errors"
	"time"

	"github.com/volatiletech/null/v8"
)

// ChatID represents unique identifier of Chat in Share File Bot.
type ChatID int

const ZeroChatID = ChatID(0)

// ChatType define types of chat.
type ChatType int8

const (
	// ChatTypeGroup represents simple group chat in Telegram.
	ChatTypeGroup ChatType = iota + 1
	// ChatTypeSuperGroup represents large group chat in Telegram.
	ChatTypeSuperGroup
	// ChatTypeChannel represents channel in Telegram.
	ChatTypeChannel
)

var (
	// ErrInvalidChatType returns when provided chat type is invalid.
	ErrInvalidChatType = errors.New("invalid chat type")
)

// ParseChatType convert string to chat type, or return error.
func ParseChatType(v string) (ChatType, error) {
	switch v {
	case "Group":
		return ChatTypeGroup, nil
	case "SuperGroup":
		return ChatTypeSuperGroup, nil
	case "Channel":
		return ChatTypeChannel, nil
	default:
		return ChatType(0), ErrInvalidChatType
	}
}

// Chat represents chat linked to Share File Bot.
type Chat struct {
	// Unique ID of chat
	ID ChatID

	// Unique ID of chat in Telegram
	TelegramID int64

	// Title of chat.
	Title string

	// Type represents type of chat.
	Type ChatType

	// OwnerID represents user who link the chat.
	OwnerID UserID

	// LinkedAt time when chat was linked to Share File Bot.
	LinkedAt time.Time

	// UpdatedAt time when chat was last updated in Share File Bot.
	UpdatedAt null.Time
}

// Patch of chat. Modify chat in do and check if something changed.
func (chat *Chat) Patch(do func(*Chat)) bool {
	newChat := *chat

	do(&newChat)

	var updated bool

	if newChat.Title != chat.Title {
		chat.Title = newChat.Title
		updated = true
	}

	if updated {
		chat.UpdatedAt = null.TimeFrom(time.Now())
	}

	return updated
}

// NewChat creates a chat by provided info
func NewChat(tgID int64, title string, typ ChatType, ownerID UserID) *Chat {
	return &Chat{
		TelegramID: tgID,
		Title:      title,
		Type:       typ,
		OwnerID:    ownerID,
		LinkedAt:   time.Now(),
	}
}

var ErrChatNotFound = errors.New("chat not found")

// ChatStore define interface for persistence of chat.
type ChatStore interface {
	// Add chat to store.
	Add(ctx context.Context, chat *Chat) error

	// Update chat in store.
	Update(ctx context.Context, chat *Chat) error

	Query() ChatStoreQuery
}

// ChatStoreQuery define interface for complex queries.
type ChatStoreQuery interface {
	// Filter Users by ID's
	ID(ids ...ChatID) ChatStoreQuery

	TelegramID(v int64) ChatStoreQuery

	// UserID filter response by user id
	OwnerID(id UserID) ChatStoreQuery

	// Query only one item from store.
	One(ctx context.Context) (*Chat, error)

	// Query all items from store.
	All(ctx context.Context) ([]*Chat, error)

	// Delete all matched objects
	Delete(ctx context.Context) (int, error)

	// Count items in store.
	Count(ctx context.Context) (int, error)
}
