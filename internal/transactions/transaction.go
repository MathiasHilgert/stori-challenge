package transactions

import "time"

// Transaction represents a financial transaction made into a bank account.
type Transaction struct {
	// ID is the unique identifier for the transaction.
	ID uint

	// Date is the date when the transaction occurred.
	Date time.Time

	// Amount is the monetary value of the transaction.
	// In this case, for demo purposes, we don't care about currency or precision.
	Amount float64

	// AccountID is the identifier of the account associated with this transaction.
	AccountID string
}
