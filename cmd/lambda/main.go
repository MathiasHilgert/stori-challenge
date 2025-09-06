// Package main wires the Lambda entrypoint for processing S3-created CSV files.
// It builds a single, reusable application processor during the cold start and
// handles each S3 record defensively. Partial failures are accounted for without
// retriggering the entire batch (see the Handler return semantics below).
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"stori-challenge/internal/application"
	"stori-challenge/internal/summaries"
	"stori-challenge/internal/summaries/mailing"
	"stori-challenge/internal/transactions"
	"stori-challenge/pkg/blend"
)

const (
	// InitializationTimeout defines the maximum time allowed for cold start initialization
	InitializationTimeout = 30 * time.Second

	// ProcessingTimeout defines the maximum time allowed for processing a single file
	ProcessingTimeout = 5 * time.Minute
)

// ApplicationDependencies encapsulates all dependencies needed by the processor.
// This makes testing easier and dependency management more explicit.
type ApplicationDependencies struct {
	Logger     blend.Logger
	Storage    summaries.SummaryFilesStorage
	Loader     transactions.TransactionLoader
	Repository transactions.TransactionsRepository
	Summarizer summaries.Summarizer
	Mailer     mailing.Mailer
	Config     application.ApplicationConfig
}

// processor is built once (cold start) and reused across invocations.
// This minimizes per-invocation latency and avoids repeated client initialization.
var (
	processor     application.TransactionProcessor
	processorOnce sync.Once
	initError     error
)

// getProcessor returns the singleton processor instance, initializing it once.
// This ensures thread-safe initialization and proper error handling.
func getProcessor() (application.TransactionProcessor, error) {
	processorOnce.Do(func() {
		processor, initError = buildProcessor()
	})
	return processor, initError
}

// buildProcessor constructs all dependencies and returns a fully wired
// TransactionProcessor. It uses structured error handling and timeouts
// for better reliability and observability.
func buildProcessor() (application.TransactionProcessor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), InitializationTimeout)
	defer cancel()

	logger, err := initializeLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info(ctx, "Starting application initialization...")

	deps, err := buildDependencies(ctx, logger)
	if err != nil {
		logger.Error(ctx, "Application initialization failed: %v", err)
		return nil, fmt.Errorf("failed to build dependencies: %w", err)
	}

	processor := application.NewProcessor(
		deps.Logger,
		deps.Storage,
		deps.Loader,
		deps.Repository,
		deps.Summarizer,
		deps.Mailer,
	)

	logger.Info(ctx, "Application initialization completed successfully")
	return processor, nil
}

// initializeLogger creates and configures the application logger.
func initializeLogger() (blend.Logger, error) {
	logger, err := blend.Default(os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return logger, nil
}

// buildDependencies constructs all application dependencies with proper error handling.
func buildDependencies(ctx context.Context, logger blend.Logger) (*ApplicationDependencies, error) {
	// 1) Load AWS config from the environment/role chain.
	logger.Debug(ctx, "Loading AWS configuration...")
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 2) Compose configuration providers.
	logger.Debug(ctx, "Initializing configuration providers...")
	envProvider := &application.DefaultEnvProvider{}
	secretsProvider := application.NewAWSSecretsProvider(secretsmanager.NewFromConfig(awsCfg))

	// 3) Load strongly-typed application configuration.
	logger.Debug(ctx, "Loading application configuration...")
	var appCfg application.ApplicationConfig
	if err := appCfg.Load(ctx, envProvider, secretsProvider); err != nil {
		return nil, fmt.Errorf("failed to load application config: %w", err)
	}

	// 4) Instantiate AWS clients once.
	logger.Debug(ctx, "Initializing AWS clients...")
	s3Client := s3.NewFromConfig(awsCfg)
	ddbClient := dynamodb.NewFromConfig(awsCfg)

	// 5) Build domain components.
	logger.Debug(ctx, "Building domain components...")
	storage := summaries.NewS3SummaryFilesStorage(s3Client)
	loader := transactions.NewCSVTransactionLoader()
	repo := transactions.NewDynamoTransactionsRepository(ddbClient, appCfg.TransactionsDynamoDB.TableName)
	summarizer := summaries.NewDefaultSummarizer()
	mailer := mailing.NewSMTPMailer(mailing.SMTPConfig(appCfg.EmailSMTP))

	return &ApplicationDependencies{
		Logger:     logger,
		Storage:    storage,
		Loader:     loader,
		Repository: repo,
		Summarizer: summarizer,
		Mailer:     mailer,
		Config:     appCfg,
	}, nil
}

// ProcessingStats tracks processing statistics for better observability.
type ProcessingStats struct {
	TotalRecords   int
	SuccessCount   int
	FailureCount   int
	ProcessingTime time.Duration
	Errors         []error
}

// AddError safely adds an error to the processing stats.
func (s *ProcessingStats) AddError(recordIndex int, bucket, key string, err error) {
	s.FailureCount++
	wrappedErr := fmt.Errorf("record %d (s3://%s/%s): %w", recordIndex, bucket, key, err)
	s.Errors = append(s.Errors, wrappedErr)
}

// Handler is the Lambda entrypoint for S3 "ObjectCreated:*" notifications.
// It iterates over each record, validates input, and processes the corresponding
// object. Failures are accumulated and reported with enhanced error handling
// and observability.
func Handler(ctx context.Context, event events.S3Event) (string, error) {
	startTime := time.Now()

	// Get the processor instance (initialized once)
	proc, err := getProcessor()
	if err != nil {
		return "", fmt.Errorf("failed to initialize processor: %w", err)
	}

	logger, _ := initializeLogger() // Safe to ignore error as getProcessor succeeded
	logger.Info(ctx, "Starting S3 event processing with %d records...", len(event.Records))

	stats := &ProcessingStats{
		TotalRecords: len(event.Records),
		Errors:       make([]error, 0),
	}

	// Process each record with individual timeout and error handling
	for i, rec := range event.Records {
		if err := processRecord(ctx, logger, proc, i, rec, stats); err != nil {
			// Error already logged and added to stats in processRecord
			continue
		}
		stats.SuccessCount++
	}

	stats.ProcessingTime = time.Since(startTime)
	return generateSummary(ctx, logger, stats), nil
}

// processRecord handles the processing of a single S3 record with proper error handling.
func processRecord(ctx context.Context, logger blend.Logger, proc application.TransactionProcessor,
	recordIndex int, rec events.S3EventRecord, stats *ProcessingStats) error {

	// Create a timeout context for this specific record
	recordCtx, cancel := context.WithTimeout(ctx, ProcessingTimeout)
	defer cancel()

	bucket, key, err := validateS3Record(rec)
	if err != nil {
		logger.Error(ctx, "Failed to validate S3 record %d: %v", recordIndex, err)
		stats.AddError(recordIndex, "", "", fmt.Errorf("validation failed: %w", err))
		return err
	}

	logger.Info(ctx, "Processing file from S3: s3://%s/%s...", bucket, key)

	// Process the file with timeout context
	if _, err := proc.ProcessFile(recordCtx, bucket, key); err != nil {
		logger.Error(ctx, "Failed to process file s3://%s/%s: %v", bucket, key, err)
		stats.AddError(recordIndex, bucket, key, err)
		return err
	}

	logger.Info(ctx, "Successfully processed file s3://%s/%s", bucket, key)
	return nil
}

// generateSummary creates a comprehensive summary of the processing results.
func generateSummary(ctx context.Context, logger blend.Logger, stats *ProcessingStats) string {
	summary := fmt.Sprintf(
		"S3 event processing completed: %d succeeded, %d failed (total: %d, duration: %v)",
		stats.SuccessCount, stats.FailureCount, stats.TotalRecords, stats.ProcessingTime,
	)

	if len(stats.Errors) > 0 {
		logger.Warn(ctx, "Processing completed with %d errors", len(stats.Errors))
		for _, err := range stats.Errors {
			logger.Error(ctx, "Error details: %v", err)
		}
	}

	logger.Info(ctx, summary)
	return summary
}

// validateS3Record ensures the record has the minimum information required.
// It provides detailed validation with specific error messages for better debugging.
func validateS3Record(rec events.S3EventRecord) (bucket string, key string, err error) {
	bucket = rec.S3.Bucket.Name
	key = rec.S3.Object.Key

	switch {
	case bucket == "" && key == "":
		return "", "", errors.New("missing both bucket name and object key")
	case bucket == "":
		return "", "", fmt.Errorf("missing bucket name for object key: %s", key)
	case key == "":
		return "", "", fmt.Errorf("missing object key for bucket: %s", bucket)
	default:
		return bucket, key, nil
	}
}

func main() {
	lambda.Start(Handler)
}
