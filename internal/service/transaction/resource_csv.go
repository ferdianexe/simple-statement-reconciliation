package transaction

import (
	"context"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

func (svc *Resource) ParseSystemTransactions(ctx context.Context, path string) ([]csv.Transaction, error) {
	return svc.csvRepo.ParseSystemTransactions(ctx, path)
}
