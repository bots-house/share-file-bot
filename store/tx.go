package store

import "context"

// StoreTx define interface of transactional of store.
type StoreTx interface {
	// TxFactory returns function for create transaction scopes.
	Tx(ctx context.Context, txFunc TxFunc) error
}

// TxFunc define signature of callback used in tx block
type TxFunc func(ctx context.Context) error

// Txier define function to start tx block
type Txier func(ctx context.Context, txFunc TxFunc) error
