package summaries

import (
	"context"
	"testing"
	"time"

	"stori-challenge/internal/transactions"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSummarizer_CalculateSummary(t *testing.T) {
	// Arrange
	summarizer := NewDefaultSummarizer()

	tests := []struct {
		name                 string
		transactions         []transactions.Transaction
		expectedTotalBalance float64
		expectedYearlyData   YearlyData
	}{
		{
			name:                 "it should return empty summary for no transactions",
			transactions:         []transactions.Transaction{},
			expectedTotalBalance: 0,
			expectedYearlyData:   YearlyData{},
		},
		{
			name: "it should calculate summary for single transaction",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: 100.50},
			},
			expectedTotalBalance: 100.50,
			expectedYearlyData: YearlyData{
				SummaryYear(2023): MonthlyData{
					time.July: MonthlySummary{
						TransactionCount: 1,
						AverageDebit:     0,
						AverageCredit:    100.50,
					},
				},
			},
		},
		{
			name: "it should calculate summary for multiple transactions in same month",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: 100.00},
				{ID: 2, Date: time.Date(2023, time.July, 20, 0, 0, 0, 0, time.UTC), Amount: -50.00},
				{ID: 3, Date: time.Date(2023, time.July, 25, 0, 0, 0, 0, time.UTC), Amount: 200.00},
			},
			expectedTotalBalance: 250.00,
			expectedYearlyData: YearlyData{
				SummaryYear(2023): MonthlyData{
					time.July: MonthlySummary{
						TransactionCount: 3,
						AverageDebit:     -50.00,
						AverageCredit:    150.00, // (100 + 200) / 2
					},
				},
			},
		},
		{
			name: "it should calculate summary for multiple transactions across different months",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: 100.00},
				{ID: 2, Date: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC), Amount: -30.00},
				{ID: 3, Date: time.Date(2023, time.August, 20, 0, 0, 0, 0, time.UTC), Amount: 75.00},
			},
			expectedTotalBalance: 145.00,
			expectedYearlyData: YearlyData{
				SummaryYear(2023): MonthlyData{
					time.July: MonthlySummary{
						TransactionCount: 1,
						AverageDebit:     0,
						AverageCredit:    100.00,
					},
					time.August: MonthlySummary{
						TransactionCount: 2,
						AverageDebit:     -30.00,
						AverageCredit:    75.00,
					},
				},
			},
		},
		{
			name: "it should calculate summary for transactions across different years",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC), Amount: 50.00},
				{ID: 2, Date: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC), Amount: -25.00},
				{ID: 3, Date: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC), Amount: 100.00},
			},
			expectedTotalBalance: 125.00,
			expectedYearlyData: YearlyData{
				SummaryYear(2022): MonthlyData{
					time.December: MonthlySummary{
						TransactionCount: 1,
						AverageDebit:     0,
						AverageCredit:    50.00,
					},
				},
				SummaryYear(2023): MonthlyData{
					time.January: MonthlySummary{
						TransactionCount: 2,
						AverageDebit:     -25.00,
						AverageCredit:    100.00,
					},
				},
			},
		},
		{
			name: "it should handle only debit transactions",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: -100.00},
				{ID: 2, Date: time.Date(2023, time.July, 20, 0, 0, 0, 0, time.UTC), Amount: -50.00},
			},
			expectedTotalBalance: -150.00,
			expectedYearlyData: YearlyData{
				SummaryYear(2023): MonthlyData{
					time.July: MonthlySummary{
						TransactionCount: 2,
						AverageDebit:     -75.00, // (-100 + -50) / 2
						AverageCredit:    0,
					},
				},
			},
		},
		{
			name: "it should handle zero amount transactions",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: 0},
				{ID: 2, Date: time.Date(2023, time.July, 20, 0, 0, 0, 0, time.UTC), Amount: 100.00},
			},
			expectedTotalBalance: 100.00,
			expectedYearlyData: YearlyData{
				SummaryYear(2023): MonthlyData{
					time.July: MonthlySummary{
						TransactionCount: 2,
						AverageDebit:     0,
						AverageCredit:    100.00,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := summarizer.CalculateSummary(context.Background(), tt.transactions)

			// Assert
			assert.Equal(t, tt.expectedTotalBalance, result.TotalBalance, "Total balance should match")
			assert.Equal(t, tt.expectedYearlyData, result.YearlyData, "Yearly data should match")
		})
	}
}

func TestDefaultSummarizer_calculateTotalBalance(t *testing.T) {
	// Arrange
	summarizer := NewDefaultSummarizer()

	tests := []struct {
		name         string
		transactions []transactions.Transaction
		expected     float64
	}{
		{
			name:         "it should return zero for empty transactions",
			transactions: []transactions.Transaction{},
			expected:     0,
		},
		{
			name: "it should sum positive amounts",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: 100.50},
				{ID: 2, Amount: 200.25},
			},
			expected: 300.75,
		},
		{
			name: "it should sum negative amounts",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: -50.00},
				{ID: 2, Amount: -25.50},
			},
			expected: -75.50,
		},
		{
			name: "it should sum mixed amounts",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: 100.00},
				{ID: 2, Amount: -30.00},
				{ID: 3, Amount: 50.00},
			},
			expected: 120.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := summarizer.calculateTotalBalance(tt.transactions)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSummarizer_separateDebitsAndCredits(t *testing.T) {
	// Arrange
	summarizer := NewDefaultSummarizer()

	tests := []struct {
		name            string
		transactions    []transactions.Transaction
		expectedDebits  []float64
		expectedCredits []float64
	}{
		{
			name:            "it should return empty slices for no transactions",
			transactions:    []transactions.Transaction{},
			expectedDebits:  nil,
			expectedCredits: nil,
		},
		{
			name: "it should separate debits and credits",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: 100.00},
				{ID: 2, Amount: -50.00},
				{ID: 3, Amount: 200.00},
				{ID: 4, Amount: -25.00},
			},
			expectedDebits:  []float64{-50.00, -25.00},
			expectedCredits: []float64{100.00, 200.00},
		},
		{
			name: "it should handle only credits",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: 100.00},
				{ID: 2, Amount: 200.00},
			},
			expectedDebits:  nil,
			expectedCredits: []float64{100.00, 200.00},
		},
		{
			name: "it should handle only debits",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: -50.00},
				{ID: 2, Amount: -75.00},
			},
			expectedDebits:  []float64{-50.00, -75.00},
			expectedCredits: nil,
		},
		{
			name: "it should ignore zero amounts",
			transactions: []transactions.Transaction{
				{ID: 1, Amount: 0},
				{ID: 2, Amount: 100.00},
				{ID: 3, Amount: -50.00},
			},
			expectedDebits:  []float64{-50.00},
			expectedCredits: []float64{100.00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			debits, credits := summarizer.separateDebitsAndCredits(tt.transactions)

			// Assert
			assert.Equal(t, tt.expectedDebits, debits, "Debits should match")
			assert.Equal(t, tt.expectedCredits, credits, "Credits should match")
		})
	}
}

func TestDefaultSummarizer_calculateAverage(t *testing.T) {
	// Arrange
	summarizer := NewDefaultSummarizer()

	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "it should return zero for empty slice",
			values:   []float64{},
			expected: 0,
		},
		{
			name:     "it should calculate average for single value",
			values:   []float64{100.00},
			expected: 100.00,
		},
		{
			name:     "it should calculate average for multiple values",
			values:   []float64{100.00, 200.00, 300.00},
			expected: 200.00,
		},
		{
			name:     "it should calculate average for negative values",
			values:   []float64{-50.00, -100.00},
			expected: -75.00,
		},
		{
			name:     "it should calculate average for mixed values",
			values:   []float64{-50.00, 100.00, 200.00},
			expected: 83.33333333333333,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := summarizer.calculateAverage(tt.values)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSummarizer_groupTransactionsByYearAndMonth(t *testing.T) {
	// Arrange
	summarizer := NewDefaultSummarizer()

	tests := []struct {
		name         string
		transactions []transactions.Transaction
		expected     map[SummaryYear]map[time.Month][]transactions.Transaction
	}{
		{
			name:         "it should return empty map for no transactions",
			transactions: []transactions.Transaction{},
			expected:     map[SummaryYear]map[time.Month][]transactions.Transaction{},
		},
		{
			name: "it should group transactions by year and month",
			transactions: []transactions.Transaction{
				{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: 100.00},
				{ID: 2, Date: time.Date(2023, time.July, 20, 0, 0, 0, 0, time.UTC), Amount: -50.00},
				{ID: 3, Date: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC), Amount: 200.00},
				{ID: 4, Date: time.Date(2024, time.July, 5, 0, 0, 0, 0, time.UTC), Amount: 75.00},
			},
			expected: map[SummaryYear]map[time.Month][]transactions.Transaction{
				SummaryYear(2023): {
					time.July: {
						{ID: 1, Date: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), Amount: 100.00},
						{ID: 2, Date: time.Date(2023, time.July, 20, 0, 0, 0, 0, time.UTC), Amount: -50.00},
					},
					time.August: {
						{ID: 3, Date: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC), Amount: 200.00},
					},
				},
				SummaryYear(2024): {
					time.July: {
						{ID: 4, Date: time.Date(2024, time.July, 5, 0, 0, 0, 0, time.UTC), Amount: 75.00},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := summarizer.groupTransactionsByYearAndMonth(tt.transactions)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewDefaultSummarizer(t *testing.T) {
	t.Run("it should create a new DefaultSummarizer instance", func(t *testing.T) {
		// Act
		summarizer := NewDefaultSummarizer()

		// Assert
		assert.NotNil(t, summarizer)
		assert.IsType(t, &DefaultSummarizer{}, summarizer)
	})
}
