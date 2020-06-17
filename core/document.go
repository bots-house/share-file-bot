package core

import (
	"context"
	"time"
)

// DocumentID it's alias for share id.
type DocumentID int

// Document represents shared document.
type Document struct {
	// Unique ID of Document.
	ID DocumentID

	// Telegram File ID
	FileID string

	// Telegram Unique File ID
	UniqueFileID string

	// Caption of file
	Caption string

	// MIMEType of file
	MIMEType string

	// Reference to user who uploads document.
	OwnerID UserID

	// Time when Document was created.
	CreatedAt time.Time
}

// DocumentStore define persistance interface for Document.
type DocumentStore interface {
	// Add Document to store. Update ID.
	Add(ctx context.Context, Document *Document) error

	// Find Document in store by ID.
	Find(ctx context.Context, id DocumentID) (*Document, error)
}
