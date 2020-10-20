package store

import (
	"github.com/bots-house/share-file-bot/core"
)

// StoreFactory define interface of factory methods
type StoreFactory interface {
	User() core.UserStore
	File() core.FileStore
	Download() core.DownloadStore
	Chat() core.ChatStore
}

// Store define generic interface for database with transaction support
type Store interface {
	StoreFactory
	StoreTx
	StoreMigrator
}
