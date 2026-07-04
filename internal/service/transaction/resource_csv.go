package transaction

import (
	"context"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

// ParseSystemTransactionsFromCSV delegates to the underlying csv repository.
// It parses system transactions CSV file.
// It accept path, start, and end time as parameters.
// It returns parsed system transactions.
// If parsing failed, it returns error.
// If parsing is successful, it returns slice of csv.Transaction.
func (svc *Resource) ParseSystemTransactionsFromCSV(ctx context.Context, path string, start, end time.Time) ([]csv.Transaction, error) {
	result, err := svc.csvRepo.ParseSystemTransactions(ctx, path, start, end)
	if err != nil {
		return []csv.Transaction{}, err
	}
	return result, nil
}
