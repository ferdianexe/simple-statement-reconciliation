package bank

import (
	"context"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

func (svc *Resource) ParseBankStatement(ctx context.Context, path string, bankName string) ([]csv.BankStatement, error) {
	return svc.csvRepo.ParseBankStatement(ctx, path, bankName)
}
