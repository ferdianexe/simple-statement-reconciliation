package transaction

import (
	"context"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

//go:generate mockgen -source=service.go -destination=mock_service.go -package=transaction

// resourceProvider provides resource methods needed for chatchannel service.
type resourceProvider interface {
	// ParseSystemTransactions parses system transactions CSV file.
	ParseSystemTransactions(ctx context.Context, path string) ([]csv.Transaction, error)
}

// Service type of transaction service.
type Service struct {
	resource resourceProvider
}

// NewService instantiates transaction service.
func NewService(resource resourceProvider) *Service {
	return &Service{resource: resource}
}
