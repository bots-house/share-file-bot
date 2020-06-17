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

type DownloadStoreQuery interface {
	DocumentID(id DocumentID)

	All(ctx context.Context) ([]*Document, error)
	Count(ctx context.Context) (int, error)
}

type DownloadStore interface {
	Add(ctx context.Context, download *Download) error
	Query() DownloadStoreQuery
}
