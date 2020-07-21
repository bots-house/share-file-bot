package core

import (
	"context"
	"errors"
	"time"

	"github.com/bots-house/share-file-bot/pkg/secretid"
	"github.com/volatiletech/null/v8"
)

// DocumentID it's alias for share id.
type DocumentID int

// Document represents shared document.
type Document struct {
	// Unique ID of Document.
	ID DocumentID

	// Telegram File ID
	FileID string

	// Public File ID
	PublicID string

	// Telegram Unique File ID
	// UniqueFileID string

	// Caption of file
	Caption null.String

	// MIMEType of file
	MIMEType null.String

	// File name
	Name string

	// File size in bytes
	Size int

	// Reference to user who uploads document.
	OwnerID UserID

	// Time when Document was created.
	CreatedAt time.Time
}

func (doc *Document) RegenPublicID() {
	doc.PublicID = secretid.Generate(secretid.IsLong(doc.PublicID))
}

func NewDocument(
	fileID string,
	caption string,
	mimeType string,
	size int,
	name string,
	ownerID UserID,
	longID bool,
) *Document {
	return &Document{
		FileID:    fileID,
		PublicID:  secretid.Generate(longID),
		Caption:   null.NewString(caption, caption != ""),
		MIMEType:  null.NewString(mimeType, mimeType != ""),
		Size:      size,
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
	}
}

var ErrDocumentNotFound = errors.New("document not found")

type DocumentStoreQuery interface {
	ID(id DocumentID) DocumentStoreQuery
	OwnerID(id UserID) DocumentStoreQuery
	PublicID(id string) DocumentStoreQuery

	One(ctx context.Context) (*Document, error)
	Delete(ctx context.Context) error
	Count(ctx context.Context) (int, error)
}

// DocumentStore define persistence interface for Document.
type DocumentStore interface {
	// Add Document to store. Update ID.
	Add(ctx context.Context, Document *Document) error

	Query() DocumentStoreQuery
}
