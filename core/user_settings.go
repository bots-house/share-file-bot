package core

import (
	"time"

	"github.com/volatiletech/null/v8"
)

// UserSettings contains user settings.
type UserSettings struct {
	// If true, bot generate super long id's for user files.
	LongIDs bool `json:"long_ids"`

	// Timestamp of last update of user settings.
	UpdatedAt null.Time `json:"updated_at"`
}

// Update check if something changed, apply changes and set new updated at.
func (settings *UserSettings) Update(patch UserSettings) {
	var updated bool

	if settings.LongIDs != patch.LongIDs {
		settings.LongIDs = patch.LongIDs
		updated = true
	}

	if updated {
		settings.UpdatedAt = null.TimeFrom(time.Now())
	}
}
