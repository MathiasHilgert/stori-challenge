package transactions

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// CSVTransactionLoader implements TransactionLoader for CSV data sources.
// It provides high-performance streaming CSV processing with minimal memory allocation.
type CSVTransactionLoader struct {
	// dateParser is cached to avoid repeated time format parsing
	dateLayout string

	// currentYear is cached to avoid repeated time.Now() calls
	currentYear int

	// csvConfig holds CSV parsing configuration
	csvConfig CSVTransactionLoaderConfig
}

// CSVTransactionLoaderConfig holds configuration for CSV parsing optimization.
type CSVTransactionLoaderConfig struct {
	// BufferSize for the CSV reader buffer (default: 64KB)
	BufferSize int

	// ExpectedRecords provides hint for slice pre-allocation
	ExpectedRecords int

	// FieldsPerRecord expected number of fields per record for validation
	FieldsPerRecord int
}

// DefaultCSVConfig returns optimized default configuration.
func DefaultCSVConfig() CSVTransactionLoaderConfig {
	return CSVTransactionLoaderConfig{
		BufferSize:      64 * 1024, // 64KB buffer for optimal I/O
		ExpectedRecords: 100,       // Reasonable default for pre-allocation
		FieldsPerRecord: 3,         // ID, Date, Transaction
	}
}

// NewCSVTransactionLoader creates a new optimized CSV transaction loader.
// Uses current year caching and optimized CSV configuration.
func NewCSVTransactionLoader() *CSVTransactionLoader {
	return NewCSVTransactionLoaderWithConfig(DefaultCSVConfig())
}

// NewCSVTransactionLoaderWithConfig creates a loader with custom configuration.
// Allows fine-tuning for specific use cases and performance requirements.
func NewCSVTransactionLoaderWithConfig(config CSVTransactionLoaderConfig) *CSVTransactionLoader {
	return &CSVTransactionLoader{
		dateLayout:  "1/2/2006",
		currentYear: time.Now().Year(),
		csvConfig:   config,
	}
}

// LoadTransactions implements streaming CSV processing with optimal memory usage.
// Uses buffered reading and context-aware processing for better performance.
func (loader *CSVTransactionLoader) LoadTransactions(ctx context.Context, reader io.Reader) ([]Transaction, error) {
	// Early context validation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error before processing: %w", err)
	}

	// Use buffered reader for better I/O performance
	bufferedReader := bufio.NewReaderSize(reader, loader.csvConfig.BufferSize)
	csvReader := csv.NewReader(bufferedReader)

	// Configure CSV reader for strict validation
	csvReader.FieldsPerRecord = loader.csvConfig.FieldsPerRecord
	csvReader.TrimLeadingSpace = true

	// Skip header row efficiently
	if _, err := csvReader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Pre-allocate slice with capacity hint for better memory efficiency
	transactions := make([]Transaction, 0, loader.csvConfig.ExpectedRecords)

	// Stream processing with minimal allocations
	lineNumber := 2 // Start from 2 (after header)
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled at line %d: %w", lineNumber, ctx.Err())
		default:
		}

		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV parsing error at line %d: %w", lineNumber, err)
		}

		transaction, err := loader.parseRecord(record, lineNumber)
		if err != nil {
			return nil, fmt.Errorf("record validation error at line %d: %w", lineNumber, err)
		}

		transactions = append(transactions, transaction)
		lineNumber++
	}

	return transactions, nil
}

// parseRecord converts a raw CSV record to Transaction with zero-allocation string processing.
// Optimized for performance with minimal string operations and direct parsing.
func (loader *CSVTransactionLoader) parseRecord(record []string, lineNumber int) (Transaction, error) {
	// Validate record length (already done by csv.Reader.FieldsPerRecord, but explicit for clarity)
	if len(record) != 3 {
		return Transaction{}, fmt.Errorf("expected 3 fields, got %d", len(record))
	}

	var transaction Transaction
	var err error

	// Parse ID with optimized error handling
	if transaction.ID, err = loader.parseIDOptimized(record[0]); err != nil {
		return Transaction{}, fmt.Errorf("invalid ID '%s': %w", record[0], err)
	}

	// Parse date with cached layout and year
	if transaction.Date, err = loader.parseDateOptimized(record[1]); err != nil {
		return Transaction{}, fmt.Errorf("invalid date '%s': %w", record[1], err)
	}

	// Parse amount with optimized float parsing
	if transaction.Amount, err = loader.parseAmountOptimized(record[2]); err != nil {
		return Transaction{}, fmt.Errorf("invalid amount '%s': %w", record[2], err)
	}

	return transaction, nil
}

// parseIDOptimized performs high-performance ID parsing with minimal allocations.
func (loader *CSVTransactionLoader) parseIDOptimized(idStr string) (uint, error) {
	// Fast path: avoid TrimSpace allocation for most cases
	if idStr == "" {
		return 0, fmt.Errorf("ID cannot be empty")
	}

	// Direct parsing without unnecessary string operations
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		// Only trim and retry if initial parsing failed
		trimmed := strings.TrimSpace(idStr)
		if trimmed == "" {
			return 0, fmt.Errorf("ID cannot be empty after trimming")
		}
		if id, err = strconv.ParseUint(trimmed, 10, 32); err != nil {
			return 0, fmt.Errorf("must be a valid positive integer")
		}
	}

	return uint(id), nil
}

// parseDateOptimized performs optimized date parsing with cached year and layout.
func (loader *CSVTransactionLoader) parseDateOptimized(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date cannot be empty")
	}

	// Use strings.Builder for efficient string concatenation
	var dateBuilder strings.Builder
	dateBuilder.Grow(len(dateStr) + 5) // Pre-allocate capacity
	dateBuilder.WriteString(dateStr)
	dateBuilder.WriteByte('/')
	dateBuilder.WriteString(strconv.Itoa(loader.currentYear))

	// Parse with cached layout
	date, err := time.Parse(loader.dateLayout, dateBuilder.String())
	if err != nil {
		return time.Time{}, fmt.Errorf("must be in M/D format: %w", err)
	}

	return date, nil
}

// parseAmountOptimized performs high-performance amount parsing with minimal string operations.
func (loader *CSVTransactionLoader) parseAmountOptimized(amountStr string) (float64, error) {
	if amountStr == "" {
		return 0, fmt.Errorf("amount cannot be empty")
	}

	// Fast path: handle positive numbers without prefix
	if amountStr[0] != '+' && amountStr[0] != '-' {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return 0, fmt.Errorf("must be a valid number: %w", err)
		}
		return amount, nil
	}

	// Handle signed numbers
	if amountStr[0] == '+' {
		// Remove leading + and parse
		amount, err := strconv.ParseFloat(amountStr[1:], 64)
		if err != nil {
			return 0, fmt.Errorf("must be a valid number: %w", err)
		}
		return amount, nil
	}

	// Handle negative numbers (ParseFloat handles this automatically)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("must be a valid number: %w", err)
	}

	return amount, nil
}
