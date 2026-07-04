package reconcile

import (
	"context"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/bank"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/transaction"
)

//go:generate mockgen -source=usecase.go -destination=usecase_mock.go -package=reconcile

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

// infraProvider provides the infrastructure primitives this usecase
// needs for date-range logic, injected so it can be mocked in tests
// rather than this layer importing gotime directly.
type infraProvider interface {
	// TimeTruncateToDay strips the time-of-day component from t.
	TimeTruncateToDay(t time.Time) time.Time
	// TimeInRange reports whether d's date falls within [start, end] inclusive.
	TimeInRange(d, start, end time.Time) bool
}

// Usecase type of usecase reconcile.
type Usecase struct {
	bank        bankServiceManager
	transaction transactionServiceManager
	infra       infraProvider
}

// NewUsecase instantiates reconcile handler.
func NewUsecase(bank bankServiceManager, transaction transactionServiceManager, infra infraProvider) *Usecase {
	return &Usecase{
		bank:        bank,
		transaction: transaction,
		infra:       infra,
	}
}
