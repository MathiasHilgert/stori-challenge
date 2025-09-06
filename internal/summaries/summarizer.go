package summaries

import (
	"context"
	"stori-challenge/internal/transactions"
	"time"
)

// Summarizer defines the interface for calculating transaction summaries.
type Summarizer interface {
	// CalculateSummary processes a slice of transactions and returns a comprehensive summary
	// including total balance and monthly aggregated data.
	CalculateSummary(ctx context.Context, transactions []transactions.Transaction) Summary
}

// DefaultSummarizer provides the default implementation of the Summarizer interface.
type DefaultSummarizer struct{}

// NewDefaultSummarizer creates a new instance of DefaultSummarizer.
func NewDefaultSummarizer() *DefaultSummarizer {
	return &DefaultSummarizer{}
}

// CalculateSummary implements the Summarizer interface by processing transactions
// and generating a comprehensive summary with total balance and yearly/monthly data.
func (ds *DefaultSummarizer) CalculateSummary(ctx context.Context, txns []transactions.Transaction) Summary {
	if len(txns) == 0 {
		return Summary{
			TotalBalance: 0,
			YearlyData:   make(YearlyData),
		}
	}

	// Calculate total balance
	totalBalance := ds.calculateTotalBalance(txns)

	// Group transactions by year and month and calculate data
	yearlyData := ds.calculateYearlyData(txns)

	return Summary{
		TotalBalance: totalBalance,
		YearlyData:   yearlyData,
	}
}

// calculateTotalBalance sums all transaction amounts to get the account balance.
func (ds *DefaultSummarizer) calculateTotalBalance(txns []transactions.Transaction) float64 {
	var total float64
	for _, txn := range txns {
		total += txn.Amount
	}
	return total
}

// calculateYearlyData groups transactions by year and month and calculates aggregated data.
func (ds *DefaultSummarizer) calculateYearlyData(txns []transactions.Transaction) YearlyData {
	yearlyGroups := ds.groupTransactionsByYearAndMonth(txns)
	return ds.computeMonthlyDataFromYearlyGroups(yearlyGroups)
}

// groupTransactionsByYearAndMonth groups transactions by year and then by month.
func (ds *DefaultSummarizer) groupTransactionsByYearAndMonth(txns []transactions.Transaction) map[SummaryYear]map[time.Month][]transactions.Transaction {
	yearlyGroups := make(map[SummaryYear]map[time.Month][]transactions.Transaction)

	for _, txn := range txns {
		year := SummaryYear(txn.Date.Year())
		month := txn.Date.Month()

		if yearlyGroups[year] == nil {
			yearlyGroups[year] = make(map[time.Month][]transactions.Transaction)
		}

		yearlyGroups[year][month] = append(yearlyGroups[year][month], txn)
	}

	return yearlyGroups
}

// computeMonthlyDataFromYearlyGroups calculates aggregated data for each month from grouped transactions.
func (ds *DefaultSummarizer) computeMonthlyDataFromYearlyGroups(yearlyGroups map[SummaryYear]map[time.Month][]transactions.Transaction) YearlyData {
	result := make(YearlyData)

	for year, monthGroups := range yearlyGroups {
		result[year] = make(MonthlyData)

		for month, monthTxns := range monthGroups {
			debits, credits := ds.separateDebitsAndCredits(monthTxns)
			avgDebit := ds.calculateAverage(debits)
			avgCredit := ds.calculateAverage(credits)

			result[year][month] = MonthlySummary{
				TransactionCount: len(monthTxns),
				AverageDebit:     avgDebit,
				AverageCredit:    avgCredit,
			}
		}
	}

	return result
}

// separateDebitsAndCredits separates transactions into debits and credits.
func (ds *DefaultSummarizer) separateDebitsAndCredits(txns []transactions.Transaction) ([]float64, []float64) {
	var debits, credits []float64
	for _, txn := range txns {
		if txn.Amount < 0 {
			debits = append(debits, txn.Amount)
		} else if txn.Amount > 0 {
			credits = append(credits, txn.Amount)
		}
	}
	return debits, credits
}

// calculateAverage calculates the average of a slice of float64 values.
// Returns 0 if the slice is empty.
func (ds *DefaultSummarizer) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}
