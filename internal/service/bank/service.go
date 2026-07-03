package bank

import (
	"context"

	csv "github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

//go:generate mockgen -source=service.go -destination=mock_service.go -package=bank

// resourceProvider provides resource methods needed for chatchannel service.
type resourceProvider interface {
	// ParseBankStatement parses bank statement CSV file.
	ParseBankStatement(ctx context.Context, path string, bankName string) ([]csv.BankStatement, error)
}

// Service type of bank service.
type Service struct {
	resource resourceProvider
}

// NewService instantiates bank service.
func NewService(resource resourceProvider) *Service {
	return &Service{resource: resource}
}
