package bank

import (
	"context"
	"time"

	csv "github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=bank

// resourceProvider provides resource methods needed for chatchannel service.
type resourceProvider interface {
	// ParseBankStatement parses bank statement CSV file.
	ParseBankStatementFromCSV(ctx context.Context, path string, bankName string, start, end time.Time) ([]csv.BankStatement, error)
}

// Service type of bank service.
type Service struct {
	resource resourceProvider
}

// NewService instantiates bank service.
func NewService(resource resourceProvider) *Service {
	return &Service{resource: resource}
}
