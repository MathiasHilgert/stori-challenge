package summaries

import "time"

// SummaryYear represents a year as an integer (e.g., 2023)
type SummaryYear uint

// MonthlyData represents a mapping of months to their aggregated data
type MonthlyData map[time.Month]MonthlySummary

// YearlyData represents a mapping of years to their monthly data
type YearlyData map[SummaryYear]MonthlyData

// MonthlySummary represents the aggregated data for transactions in a specific month.
type MonthlySummary struct {
	// TransactionCount is the total number of transactions in this month
	TransactionCount int

	// AverageDebit is the average amount of debit transactions in this month
	// Returns 0 if there are no debit transactions
	AverageDebit float64

	// AverageCredit is the average amount of credit transactions in this month
	// Returns 0 if there are no credit transactions
	AverageCredit float64
}

// Summary represents the complete summary of account transactions.
type Summary struct {
	// TotalBalance is the sum of all transaction amounts
	TotalBalance float64

	// YearlyData contains aggregated data grouped by year and then by month
	YearlyData YearlyData
}
