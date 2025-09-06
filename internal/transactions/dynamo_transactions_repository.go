package transactions

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// DynamoTransactionsRepository implements the TransactionsRepository interface
// using AWS DynamoDB as the persistent storage backend.
type DynamoTransactionsRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoTransactionsRepository creates a new instance of DynamoTransactionsRepository.
func NewDynamoTransactionsRepository(client *dynamodb.Client, tableName string) *DynamoTransactionsRepository {
	return &DynamoTransactionsRepository{
		client:    client,
		tableName: tableName,
	}
}

// DynamoTransaction represents the structure of a transaction as stored in DynamoDB.
type DynamoTransaction struct {
	ID         string  `dynamodbav:"id"`
	InternalID uint    `dynamodbav:"internal_id"`
	Date       string  `dynamodbav:"date"`
	Amount     float64 `dynamodbav:"amount"`
	AccountID  string  `dynamodbav:"account_id"`
}

// Save persists the given transactions to DynamoDB.
// It uses batch write operations for efficiency when saving multiple transactions.
func (r *DynamoTransactionsRepository) Save(ctx context.Context, transactions []Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// DynamoDB BatchWriteItem has a limit of 25 items per request
	const batchSize = 25

	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		batch := transactions[i:end]
		if err := r.saveBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to save batch starting at index %d: %w", i, err)
		}
	}

	return nil
}

// saveBatch saves a batch of transactions using DynamoDB BatchWriteItem.
func (r *DynamoTransactionsRepository) saveBatch(ctx context.Context, transactions []Transaction) error {
	writeRequests := make([]types.WriteRequest, 0, len(transactions))

	for _, transaction := range transactions {
		// Generate UUID v4 for primary key
		primaryID := uuid.New().String()

		// Convert Transaction to DynamoTransaction
		dynamoTx := DynamoTransaction{
			ID:         primaryID,      // UUID v4 as primary key
			InternalID: transaction.ID, // Original numeric ID
			Date:       transaction.Date.Format("2006-01-02T15:04:05Z"),
			Amount:     transaction.Amount,
			AccountID:  transaction.AccountID,
		}

		// Marshal to DynamoDB attribute values
		item, err := attributevalue.MarshalMap(dynamoTx)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction %d: %w", transaction.ID, err)
		}

		writeRequest := types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		}

		writeRequests = append(writeRequests, writeRequest)
	}

	// Execute batch write request
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			r.tableName: writeRequests,
		},
	}

	result, err := r.client.BatchWriteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute batch write: %w", err)
	}

	// Handle unprocessed items (DynamoDB may not process all items in one request)
	if len(result.UnprocessedItems) > 0 {
		return r.handleUnprocessedItems(ctx, result.UnprocessedItems)
	}

	return nil
}

// handleUnprocessedItems retries unprocessed items from a batch write operation.
func (r *DynamoTransactionsRepository) handleUnprocessedItems(ctx context.Context, unprocessedItems map[string][]types.WriteRequest) error {
	maxRetries := 3
	retryCount := 0

	for len(unprocessedItems) > 0 && retryCount < maxRetries {
		retryCount++

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: unprocessedItems,
		}

		result, err := r.client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to retry unprocessed items (attempt %d): %w", retryCount, err)
		}

		unprocessedItems = result.UnprocessedItems
	}

	if len(unprocessedItems) > 0 {
		return fmt.Errorf("failed to process all items after %d retries, %d items remain unprocessed", maxRetries, len(unprocessedItems[r.tableName]))
	}

	return nil
}
