package transaction

import (
	"context"
	"fmt"
)

// GetUserTransactionHistory loads the internal system's transaction
// history for request.SysPath within [StartDate, EndDate], and
// translates it into the service-layer Transaction type.
func (svc *Service) GetUserTransactionHistory(ctx context.Context, request TransactionHistoryParams) ([]Transaction, error) {
	sysTrx, err := svc.resource.ParseSystemTransactionsFromCSV(ctx, request.SysPath, request.StartDate, request.EndDate)
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
