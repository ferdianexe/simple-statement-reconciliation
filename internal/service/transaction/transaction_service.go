package transaction

import (
	"context"
	"fmt"
)

func (svc *Service) GetUserTransactionHistory(ctx context.Context, request TransactionHistoryParams) ([]Transaction, error) {
	sysTrx, err := svc.resource.ParseSystemTransactions(ctx, request.SysPath)
	if err != nil {
		return []Transaction{}, fmt.Errorf("parsing system transactions: %w", err)
	}

	svcTrx := make([]Transaction, 0, len(sysTrx))
	for _, t := range sysTrx {
		svcTrx = append(svcTrx, Transaction{
			TrxID:           t.TrxID,
			Amount:          t.Amount,
			Type:            TrxType(t.Type),
			TransactionTime: t.TransactionTime,
		})
	}
	return svcTrx, nil
}
