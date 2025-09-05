package transactions

import (
	"context"
	"io"
)

// TransactionLoader defines the interface for loading transactions from a data source.
type TransactionLoader interface {
	// LoadTransactions loads transactions from a data source.
	LoadTransactions(ctx context.Context, reader io.Reader) ([]Transaction, error)
}
