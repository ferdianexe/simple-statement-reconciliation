package bank

import (
	"context"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

// ParseBankStatementFromCSV resource function for repository implementation.
// it accept path, bankName, start, end time.Time as parameter.
// it returns []csv.BankStatement and error.
// if error is nil, then the function is successful.
// if error is not nil, then the function is failed.
func (svc *Resource) ParseBankStatementFromCSV(ctx context.Context, path string, bankName string, start, end time.Time) ([]csv.BankStatement, error) {
	result, err := svc.csvRepo.ParseBankStatement(ctx, path, bankName, start, end)
	if err != nil {
		return []csv.BankStatement{}, err
	}
	return result, nil
}
