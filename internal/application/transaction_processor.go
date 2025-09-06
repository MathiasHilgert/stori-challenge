package application

import (
	"context"
	"fmt"
	"stori-challenge/internal/summaries"
	"stori-challenge/internal/summaries/mailing"
	"stori-challenge/internal/transactions"
	"stori-challenge/pkg/blend"
)

// TransactionProcessor defines the pipeline contract.
type TransactionProcessor interface {
	ProcessFile(ctx context.Context, bucket, key string) (*ProcessingResult, error)
}

// DefaultProcessor implements TransactionProcessor.
type DefaultProcessor struct {
	storage    summaries.SummaryFilesStorage
	loader     transactions.TransactionLoader
	repository transactions.TransactionsRepository
	summarizer summaries.Summarizer
	mailer     mailing.Mailer
	logger     blend.Logger
}

// NewProcessor creates a new DefaultProcessor instance.
func NewProcessor(
	logger blend.Logger,
	storage summaries.SummaryFilesStorage,
	loader transactions.TransactionLoader,
	repository transactions.TransactionsRepository,
	summarizer summaries.Summarizer,
	mailer mailing.Mailer,
) *DefaultProcessor {
	return &DefaultProcessor{
		logger:     logger,
		storage:    storage,
		loader:     loader,
		repository: repository,
		summarizer: summarizer,
		mailer:     mailer,
	}
}

// ProcessingResult contains the outcome of one file.
type ProcessingResult struct {
	FilePath         string
	AccountID        string
	AccountEmail     string
	TransactionCount int
	Summary          summaries.Summary
}

// ProcessFile executes the entire pipeline strictly.
// Any failure aborts processing with an error.
func (tp *DefaultProcessor) ProcessFile(ctx context.Context, bucket, key string) (*ProcessingResult, error) {
	path := fmt.Sprintf("s3://%s/%s", bucket, key)

	// Obtain the summary file
	tp.logger.Info(ctx, "Obtaining summary file content from %s...", path)
	summaryFile, err := tp.storage.Get(ctx, path)
	if err != nil {
		tp.logger.Error(ctx, "Failed to load summary file: %v", err)
		return nil, fmt.Errorf("failed to load file: %w", err)
	}
	tp.logger.Info(ctx, "Successfully loaded summary file for account %s", summaryFile.AccountID)

	// Parse transactions
	tp.logger.Info(ctx, "Parsing transactions...")
	txns, err := tp.loader.LoadTransactions(ctx, summaryFile.Content)
	if err != nil {
		tp.logger.Error(ctx, "Failed to parse transactions: %v", err)
		return nil, fmt.Errorf("failed to parse transactions: %w", err)
	}
	for i := range txns {
		txns[i].AccountID = summaryFile.AccountID
	}
	tp.logger.Info(ctx, "Transactions parsed successfully (%d transactions)", len(txns))

	// Persist transactions
	tp.logger.Info(ctx, "Persisting transactions to repository...")
	if err := tp.repository.Save(ctx, txns); err != nil {
		tp.logger.Error(ctx, "Failed to persist transactions: %v", err)
		return nil, fmt.Errorf("failed to persist transactions: %w", err)
	}
	tp.logger.Info(ctx, "Successfully persisted %d transactions", len(txns))

	// Calculate summary
	tp.logger.Info(ctx, "Calculating summary...")
	summaryData := tp.summarizer.CalculateSummary(ctx, txns)
	tp.logger.Info(ctx, "Calculated summary for account: $(%s)", summaryFile.AccountID)

	// Send email if address is provided
	if summaryFile.AccountEmail != "" {
		tp.logger.Info(ctx, "Sending summary email to %s...", summaryFile.AccountEmail)
		if err := tp.mailer.Send(ctx, summaryFile.AccountEmail, summaryData); err != nil {
			tp.logger.Error(ctx, "Failed to send email: %v", err)
			return nil, fmt.Errorf("failed to send email: %w", err)
		}
		tp.logger.Info(ctx, "Sent summary email to %s", summaryFile.AccountEmail)
	} else {
		tp.logger.Info(ctx, "No account email provided; skipping email sending...")
	}

	// Successfully processed
	tp.logger.Info(ctx, "File %s processed successfully", path)
	return &ProcessingResult{
		FilePath:         summaryFile.Path,
		AccountID:        summaryFile.AccountID,
		AccountEmail:     summaryFile.AccountEmail,
		TransactionCount: len(txns),
		Summary:          summaryData,
	}, nil
}
