package transactions

import "context"

// TransactionsRepository is an abstraction layer to save transactions to a
// persistent storage.
type TransactionsRepository interface {
	// Save persists the given transactions to a persistent storage.
	Save(ctx context.Context, transactions []Transaction) (err error)
}
