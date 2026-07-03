package reconcile

import (
	"context"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/bank"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/transaction"
)

//go:generate mockgen -source=usecase.go -destination=mock_usecase.go -package=reconcile

// reconcileUsecaseManager provides the required action usecase methods.
type bankServiceManager interface {
	// Reconcile performs reconciliation process.
	GetBankStatementHistory(ctx context.Context, request []bank.BankStatementParams) ([]bank.BankStatement, error)
}

// transactionServiceManager provides the required action usecase methods.
type transactionServiceManager interface {
	// Reconcile performs reconciliation process.
	GetUserTransactionHistory(ctx context.Context, request transaction.TransactionHistoryParams) ([]transaction.Transaction, error)
}

// Usecase type of usecase reconcile.
type Usecase struct {
	bank        bankServiceManager
	transaction transactionServiceManager
}

// NewUsecase instantiates reconcile handler.
func NewUsecase(bank bankServiceManager, transaction transactionServiceManager) *Usecase {
	return &Usecase{
		bank:        bank,
		transaction: transaction,
	}
}
