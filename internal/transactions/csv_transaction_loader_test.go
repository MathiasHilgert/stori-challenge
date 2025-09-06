package transactions

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVTransactionLoader_LoadTransactions(t *testing.T) {
	// Arrange
	currentYear := time.Now().Year()

	testCases := []struct {
		name           string
		csvContent     string
		expectedResult []Transaction
		expectedError  string
		description    string
	}{
		{
			name: "it should successfully load valid transactions",
			csvContent: `ID,Date,Transaction
1,7/15,+60.5
2,7/28,-10.3
3,8/2,-20.46`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(currentYear, 7, 15, 0, 0, 0, 0, time.UTC),
					Amount: 60.5,
				},
				{
					ID:     2,
					Date:   time.Date(currentYear, 7, 28, 0, 0, 0, 0, time.UTC),
					Amount: -10.3,
				},
				{
					ID:     3,
					Date:   time.Date(currentYear, 8, 2, 0, 0, 0, 0, time.UTC),
					Amount: -20.46,
				},
			},
			description: "should parse valid CSV with positive and negative amounts",
		},
		{
			name: "it should successfully load transactions with full year dates",
			csvContent: `ID,Date,Transaction
1,1/5/2022,+1250.0
2,2/14/2023,-85.5
3,12/31/2024,+300.0`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC),
					Amount: 1250.0,
				},
				{
					ID:     2,
					Date:   time.Date(2023, 2, 14, 0, 0, 0, 0, time.UTC),
					Amount: -85.5,
				},
				{
					ID:     3,
					Date:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					Amount: 300.0,
				},
			},
			description: "should parse valid CSV with full year dates",
		},
		{
			name: "it should handle mixed date formats",
			csvContent: `ID,Date,Transaction
1,7/15,+60.5
2,1/5/2022,+1250.0
3,8/2,-20.46`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(currentYear, 7, 15, 0, 0, 0, 0, time.UTC),
					Amount: 60.5,
				},
				{
					ID:     2,
					Date:   time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC),
					Amount: 1250.0,
				},
				{
					ID:     3,
					Date:   time.Date(currentYear, 8, 2, 0, 0, 0, 0, time.UTC),
					Amount: -20.46,
				},
			},
			description: "should handle mixed M/D and M/D/YYYY formats in same file",
		},
		{
			name: "it should handle transactions without sign prefix",
			csvContent: `ID,Date,Transaction
1,1/1,100.00
2,2/14,250.75`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC),
					Amount: 100.00,
				},
				{
					ID:     2,
					Date:   time.Date(currentYear, 2, 14, 0, 0, 0, 0, time.UTC),
					Amount: 250.75,
				},
			},
			description: "should handle amounts without explicit + sign",
		},
		{
			name:           "it should handle empty CSV with only headers",
			csvContent:     `ID,Date,Transaction`,
			expectedResult: []Transaction{},
			description:    "should return empty slice for CSV with only headers",
		},
		{
			name: "it should reject invalid ID values",
			csvContent: `ID,Date,Transaction
abc,7/15,+60.5`,
			expectedError: "record validation error at line 2",
			description:   "should fail when ID is not a valid integer",
		},
		{
			name: "it should reject empty ID values",
			csvContent: `ID,Date,Transaction
,7/15,+60.5`,
			expectedError: "record validation error at line 2",
			description:   "should fail when ID is empty",
		},
		{
			name: "it should reject invalid date format",
			csvContent: `ID,Date,Transaction
1,2023/07/15,+60.5`,
			expectedError: "record validation error at line 2",
			description:   "should fail when date format uses slashes in wrong positions",
		},
		{
			name: "it should reject invalid date format with too many slashes",
			csvContent: `ID,Date,Transaction
1,1/2/3/4,+60.5`,
			expectedError: "record validation error at line 2",
			description:   "should fail when date has too many slash separators",
		},
		{
			name: "it should reject invalid date format with no slashes",
			csvContent: `ID,Date,Transaction
1,20230715,+60.5`,
			expectedError: "record validation error at line 2",
			description:   "should fail when date has no slash separators",
		},
		{
			name: "it should reject empty date values",
			csvContent: `ID,Date,Transaction
1,,+60.5`,
			expectedError: "record validation error at line 2",
			description:   "should fail when date is empty",
		},
		{
			name: "it should reject invalid amount values",
			csvContent: `ID,Date,Transaction
1,7/15,abc`,
			expectedError: "record validation error at line 2",
			description:   "should fail when amount is not a valid number",
		},
		{
			name: "it should reject empty amount values",
			csvContent: `ID,Date,Transaction
1,7/15,`,
			expectedError: "record validation error at line 2",
			description:   "should fail when amount is empty",
		},
		{
			name: "it should reject records with wrong number of fields",
			csvContent: `ID,Date,Transaction
1,7/15`,
			expectedError: "CSV parsing error at line 2",
			description:   "should fail when record has insufficient fields",
		},
		{
			name: "it should reject records with too many fields",
			csvContent: `ID,Date,Transaction
1,7/15,+60.5,extra`,
			expectedError: "CSV parsing error at line 2",
			description:   "should fail when record has too many fields",
		},
		{
			name:           "it should treat first line as header and skip it",
			csvContent:     `1,7/15,+60.5`,
			expectedResult: []Transaction{},
			description:    "should skip first line as header even if it contains data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			loader := NewCSVTransactionLoader()
			reader := strings.NewReader(tc.csvContent)
			ctx := context.Background()

			// Act
			result, err := loader.LoadTransactions(ctx, reader)

			// Assert
			if tc.expectedError != "" {
				assert.Error(t, err, tc.description)
				assert.Contains(t, err.Error(), tc.expectedError, tc.description)
				assert.Nil(t, result, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.Equal(t, tc.expectedResult, result, tc.description)
			}
		})
	}
}

func TestCSVTransactionLoader_LoadTransactions_ContextCancellation(t *testing.T) {
	t.Run("it should handle context cancellation before processing", func(t *testing.T) {
		// Arrange
		loader := NewCSVTransactionLoader()
		csvContent := `ID,Date,Transaction
1,7/15,+60.5`
		reader := strings.NewReader(csvContent)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Act
		result, err := loader.LoadTransactions(ctx, reader)

		// Assert
		assert.Error(t, err, "should fail when context is already cancelled")
		assert.Contains(t, err.Error(), "context", "should mention context in error")
		assert.Nil(t, result, "should return nil result on context error")
	})
}

func TestCSVTransactionLoader_CustomConfiguration(t *testing.T) {
	testCases := []struct {
		name        string
		config      CSVTransactionLoaderConfig
		csvContent  string
		expectError bool
		description string
	}{
		{
			name: "it should work with custom buffer size",
			config: CSVTransactionLoaderConfig{
				BufferSize:      1024,
				ExpectedRecords: 10,
				FieldsPerRecord: 3,
			},
			csvContent: `ID,Date,Transaction
1,7/15,+60.5`,
			expectError: false,
			description: "should work with smaller buffer size",
		},
		{
			name: "it should work with larger expected records",
			config: CSVTransactionLoaderConfig{
				BufferSize:      64 * 1024,
				ExpectedRecords: 1000,
				FieldsPerRecord: 3,
			},
			csvContent: `ID,Date,Transaction
1,7/15,+60.5`,
			expectError: false,
			description: "should work with larger expected records for pre-allocation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			loader := NewCSVTransactionLoaderWithConfig(tc.config)
			reader := strings.NewReader(tc.csvContent)
			ctx := context.Background()

			// Act
			result, err := loader.LoadTransactions(ctx, reader)

			// Assert
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotNil(t, result, tc.description)
			}
		})
	}
}

func TestCSVTransactionLoader_DefaultConfiguration(t *testing.T) {
	t.Run("it should create loader with default configuration", func(t *testing.T) {
		// Arrange & Act
		loader := NewCSVTransactionLoader()

		// Assert
		assert.NotNil(t, loader, "should create loader instance")
		assert.Equal(t, "1/2/2006", loader.dateLayout, "should use correct date layout")
		assert.Equal(t, time.Now().Year(), loader.currentYear, "should cache current year")
		assert.Equal(t, DefaultCSVConfig(), loader.csvConfig, "should use default config")
	})
}

func TestDefaultCSVConfig(t *testing.T) {
	t.Run("it should return correct default configuration", func(t *testing.T) {
		// Arrange
		expectedConfig := CSVTransactionLoaderConfig{
			BufferSize:      64 * 1024,
			ExpectedRecords: 100,
			FieldsPerRecord: 3,
		}

		// Act
		config := DefaultCSVConfig()

		// Assert
		assert.Equal(t, expectedConfig, config, "should return correct default values")
	})
}

func TestCSVTransactionLoader_EdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		csvContent     string
		expectedResult []Transaction
		expectedError  string
		description    string
	}{
		{
			name: "it should handle leading whitespace in fields",
			csvContent: `ID,Date,Transaction
 1, 7/15, +60.5`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(time.Now().Year(), 7, 15, 0, 0, 0, 0, time.UTC),
					Amount: 60.5,
				},
			},
			description: "should trim leading whitespace from fields",
		},
		{
			name: "it should handle zero amounts",
			csvContent: `ID,Date,Transaction
1,1/1,0
2,1/2,+0
3,1/3,-0`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC),
					Amount: 0,
				},
				{
					ID:     2,
					Date:   time.Date(time.Now().Year(), 1, 2, 0, 0, 0, 0, time.UTC),
					Amount: 0,
				},
				{
					ID:     3,
					Date:   time.Date(time.Now().Year(), 1, 3, 0, 0, 0, 0, time.UTC),
					Amount: 0,
				},
			},
			description: "should handle zero amounts with different signs",
		},
		{
			name: "it should handle large transaction IDs",
			csvContent: `ID,Date,Transaction
4294967295,1/1,100.00`,
			expectedResult: []Transaction{
				{
					ID:     4294967295,
					Date:   time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC),
					Amount: 100.00,
				},
			},
			description: "should handle maximum uint32 ID values",
		},
		{
			name: "it should handle decimal amounts with many digits",
			csvContent: `ID,Date,Transaction
1,1/1,123.456789`,
			expectedResult: []Transaction{
				{
					ID:     1,
					Date:   time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC),
					Amount: 123.456789,
				},
			},
			description: "should handle high precision decimal amounts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			loader := NewCSVTransactionLoader()
			reader := strings.NewReader(tc.csvContent)
			ctx := context.Background()

			// Act
			result, err := loader.LoadTransactions(ctx, reader)

			// Assert
			if tc.expectedError != "" {
				assert.Error(t, err, tc.description)
				assert.Contains(t, err.Error(), tc.expectedError, tc.description)
				assert.Nil(t, result, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				assert.Equal(t, tc.expectedResult, result, tc.description)
			}
		})
	}
}
