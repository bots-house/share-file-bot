package core

import (
	"context"
	"errors"
	"time"

	"github.com/volatiletech/null"
)

// UserID it's alias for user identifier
type UserID int

// User of bot.
type User struct {
	// Unique ID of user in bot and Telegram.
	ID int

	// First name of user from Telegram
	FirstName string

	// Last name of user from Telegram (optional)
	LastName null.String

	// Username of user from Telegram (optional)
	Username null.String

	// Language code of user from Telegram (optional)
	LanguageCode string

	// True, if user is admin of bot.
	IsAdmin bool

	// Time of first interaction with bot
	JoinedAt time.Time

	// Time when user info was updated
	UpdatedAt null.Time
}

var ErrUserNotFound = errors.New("user not found")

// UserStore define interface for persistance of bot.
type UserStore interface {
	Add(ctx context.Context, user *User) error
	Find(ctx context.Context, id UserID) (*User, error)
}
