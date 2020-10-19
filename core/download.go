package core

import (
	"context"
	"time"
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

	// At time when download was happen
	At time.Time
}

func NewDownload(fileID FileID, userID UserID) *Download {
	return &Download{
		FileID: fileID,
		UserID: userID,
		At:     time.Now(),
	}
}

type DownloadStats struct {
	Total  int
	Unique int
}

type DownloadStoreQuery interface {
	FileID(id FileID) DownloadStoreQuery

	Count(ctx context.Context) (int, error)
}

type DownloadStore interface {
	Add(ctx context.Context, download *Download) error
	GetDownloadStats(ctx context.Context, id FileID) (*DownloadStats, error)

	Query() DownloadStoreQuery
}
