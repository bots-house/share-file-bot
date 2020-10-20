package core

import (
	"context"
	"errors"
	"time"

	"github.com/volatiletech/null/v8"
)

// ChatID represents unique identifier of Chat in Share File Bot.
type ChatID int

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

// ChatStore define interface for persistance of chat.
type ChatStore interface {
	// Add chat to store.
	Add(ctx context.Context, chat *Chat) error

	// Update chat in store.
	Update(ctx context.Context, chat *Chat) error
}
