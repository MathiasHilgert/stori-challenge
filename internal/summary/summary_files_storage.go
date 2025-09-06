package summary

import (
	"context"
)

// SummaryFilesStorage defines the interface for accessing transaction files
// from a storage system. It retrieves the complete file with content and metadata.
type SummaryFilesStorage interface {
	// Get retrieves a complete SummaryFile from storage using the file path.
	// The path should contain all necessary information to locate and retrieve the file.
	// Returns a SummaryFile with content and metadata, and any error encountered.
	Get(ctx context.Context, path string) (*SummaryFile, error)
}
