package core

import (
	"context"
	"errors"
	"time"

	"github.com/volatiletech/null/v8"
)

// UserID it's alias for user identifier
type UserID int

// User of bot.
type User struct {
	// Unique ID of user in bot and Telegram.
	ID UserID

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

	// Settings of user
	Settings UserSettings

	// Ref is set we user /start with deep-link like ref_*
	Ref null.String

	// Time of first interaction with bot
	JoinedAt time.Time

	// Time when user info was updated
	UpdatedAt null.Time
}

type SummaryRefStats struct {
	ClickedOnStart int
	ConnectedChat  int
	UploadedFile   int
	SetRestriction int
	DownloadCount  int
}

func (user *User) Patch(do func(*User)) bool {
	newUser := *user

	do(&newUser)

	var updated bool

	if user.FirstName != newUser.FirstName {
		user.FirstName = newUser.FirstName
		updated = true
	}

	if user.LastName != newUser.LastName {
		user.LastName = newUser.LastName
		updated = true
	}

	if user.Username != newUser.Username {
		user.Username = newUser.Username
		updated = true
	}

	if user.Settings != newUser.Settings {
		user.Settings = newUser.Settings
		updated = true
	}

	if updated {
		user.UpdatedAt = null.TimeFrom(time.Now())
	}

	return updated
}

func NewUser(
	id UserID,
	firstName, lastName, username, langCode string,
) *User {
	return &User{
		ID:           id,
		FirstName:    firstName,
		LastName:     null.NewString(lastName, lastName != ""),
		Username:     null.NewString(username, username != ""),
		LanguageCode: langCode,
		IsAdmin:      false,
		JoinedAt:     time.Now(),
	}
}

var ErrUserNotFound = errors.New("user not found")

type UserStoreQuery interface {
	Count(ctx context.Context) (int, error)
}

type UserRefStatsItem struct {
	Ref   null.String
	Count int
}

type UserRefStats []UserRefStatsItem

// UserStore define interface for persistence of bot.
type UserStore interface {
	Add(ctx context.Context, user *User) error
	Find(ctx context.Context, id UserID) (*User, error)
	Update(ctx context.Context, user *User) error

	RefStats(ctx context.Context) (UserRefStats, error)

	SummaryRefStats(ctx context.Context, ref string) (*SummaryRefStats, error)

	Query() UserStoreQuery
}
