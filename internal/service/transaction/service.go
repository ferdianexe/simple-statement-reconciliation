package transaction

import (
	"context"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=transaction

// resourceProvider provides resource methods needed for transaction service.
type resourceProvider interface {
	// ParseSystemTransactionsFromCSV parses system transactions CSV file.
	ParseSystemTransactionsFromCSV(ctx context.Context, path string, start, end time.Time) ([]csv.Transaction, error)
}

// Service type of transaction service.
type Service struct {
	resource resourceProvider
}

// NewService instantiates transaction service.
func NewService(resource resourceProvider) *Service {
	return &Service{resource: resource}
}
