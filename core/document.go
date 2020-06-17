package core

import (
	"context"
	"errors"
	"time"

	"github.com/volatiletech/null"
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
	// UniqueFileID string

	// Caption of file
	Caption null.String

	// MIMEType of file
	MIMEType null.String

	// File size in bytes
	Size int

	// Reference to user who uploads document.
	OwnerID UserID

	// Time when Document was created.
	CreatedAt time.Time
}

var ErrDocumentNotFound = errors.New("document not found")

// DocumentStore define persistance interface for Document.
type DocumentStore interface {
	// Add Document to store. Update ID.
	Add(ctx context.Context, Document *Document) error

	// Find Document in store by ID.
	Find(ctx context.Context, id DocumentID) (*Document, error)
}
