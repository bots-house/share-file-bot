package core

import (
	"context"
	"errors"
	"time"

	"github.com/bots-house/share-file-bot/pkg/secretid"
	"github.com/volatiletech/null/v8"
)

// FileID it's alias for share id.
type FileID int

// File represents shared file.
type File struct {
	// Unique ID of File.
	ID FileID

	// Telegram File ID
	TelegramID string

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

	// Reference to user who uploads file.
	OwnerID UserID

	// Time when file was created.
	CreatedAt time.Time
}

func (doc *File) RegenPublicID() {
	doc.PublicID = secretid.Generate(secretid.IsLong(doc.PublicID))
}

func NewFile(
	fileID string,
	caption string,
	mimeType string,
	size int,
	name string,
	ownerID UserID,
	longID bool,
) *File {
	return &File{
		TelegramID: fileID,
		PublicID:   secretid.Generate(longID),
		Caption:    null.NewString(caption, caption != ""),
		MIMEType:   null.NewString(mimeType, mimeType != ""),
		Size:       size,
		Name:       name,
		OwnerID:    ownerID,
		CreatedAt:  time.Now(),
	}
}

var ErrFileNotFound = errors.New("file not found")

type FileStoreQuery interface {
	ID(id FileID) FileStoreQuery
	OwnerID(id UserID) FileStoreQuery
	PublicID(id string) FileStoreQuery

	One(ctx context.Context) (*File, error)
	Delete(ctx context.Context) error
	Count(ctx context.Context) (int, error)
}

// FileStore define persistence interface for File.
type FileStore interface {
	// Add File to store. Update ID.
	Add(ctx context.Context, file *File) error

	Query() FileStoreQuery
}
