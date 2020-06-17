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

	// Reference to document.
	DocumentID DocumentID

	// Refernce to user.
	UserID UserID

	// At time when download was happen
	At time.Time
}

func NewDownload(docID DocumentID, userID UserID) *Download {
	return &Download{
		DocumentID: docID,
		UserID:     userID,
		At:         time.Now(),
	}
}

type DownloadStats struct {
	Total  int
	Unique int
}

type DownloadStoreQuery interface {
	DocumentID(id DocumentID) DownloadStoreQuery

	Count(ctx context.Context) (int, error)
}

type DownloadStore interface {
	Add(ctx context.Context, download *Download) error
	GetDownloadStats(ctx context.Context, id DocumentID) (*DownloadStats, error)

	Query() DownloadStoreQuery
}
