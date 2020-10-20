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

// Patch check if something changed, apply changes and set new updated at.
func (settings *UserSettings) Patch(do func(*UserSettings)) bool {
	newSettings := *settings

	do(&newSettings)

	var updated bool

	if newSettings.LongIDs != settings.LongIDs {
		settings.LongIDs = newSettings.LongIDs
		updated = true
	}

	if updated {
		settings.UpdatedAt = null.TimeFrom(time.Now())
	}

	return updated
}
