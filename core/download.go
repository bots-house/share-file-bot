package core

import (
	"context"
	"time"

	"github.com/volatiletech/null/v8"
)

// DownloadID alias for download.
type DownloadID int

type Download struct {
	// ID of download.
	ID DownloadID

	// Reference to file. Can be null.
	FileID FileID

	// References to user. Can be null.
	UserID UserID

	// If true, means user was requested to subscription and successefuly subscribed,
	// False means, user was already subscribed,
	// Null means check is disable.
	NewSubscription null.Bool

	// At time when download was happen
	At time.Time
}

func (dwn *Download) SetNewSubscription(v bool) {
	dwn.NewSubscription = null.NewBool(v, true)
}

func NewDownload(fileID FileID, userID UserID) *Download {
	return &Download{
		FileID: fileID,
		UserID: userID,
		At:     time.Now(),
	}
}

// DownloadStats of file.
type DownloadStats struct {
	// Total downloads count
	Total int

	// Unique downloads count
	Unique int

	// Downloads with subscription
	WithSubscription int

	// Downloads with new subscription
	NewSubscription int
}

type DownloadStoreQuery interface {
	FileID(id FileID) DownloadStoreQuery

	Count(ctx context.Context) (int, error)
}

type DownloadStore interface {
	Add(ctx context.Context, download *Download) error
	GetFileDownloadStats(ctx context.Context, id FileID) (*DownloadStats, error)

	Query() DownloadStoreQuery
}
