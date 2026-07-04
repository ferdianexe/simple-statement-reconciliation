package bank

import (
	"context"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
)

//go:generate mockgen -source=resource.go -destination=resource_mock.go -package=bank

// csvRepoProvider is the interface that wraps the csv repository methods needed for resource object.
type csvRepoProvider interface {
	// ParseBankStatement reads one bank statement CSV file. bankName tags every
	// record so unmatched results can later be grouped per bank, since the
	// service supports reconciling against several banks in a single run.
	ParseBankStatement(ctx context.Context, path string, bankName string, start, end time.Time) ([]csv.BankStatement, error)
}

// Resource type of a action resource.
// It contains the repositories used for this resource.
// It encapsulates methods from repository and translates it into service.
type Resource struct {
	csvRepo csvRepoProvider
}

// NewResource initiates new bank resource.
func NewResource(csvRepo csvRepoProvider) *Resource {
	return &Resource{
		csvRepo: csvRepo,
	}
}
