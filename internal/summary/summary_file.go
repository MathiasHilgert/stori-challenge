package summary

import "io"

// SummaryFile represents a transaction summary file with its content and metadata.
// This file contains transaction data along with associated account information.
type SummaryFile struct {
	// Path is the full path to the file in storage (e.g., "s3://bucket/key").
	Path string

	// AccountID is the unique identifier for the account associated with the transactions.
	AccountID string

	// AccountEmail is the email address associated with the account.
	AccountEmail string

	// Content is a reader for the file's content.
	Content io.Reader
}
