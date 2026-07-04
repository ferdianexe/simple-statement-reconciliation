package csv

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
)

//go:generate mockgen -source=repo.go -destination=repo_mock.go -package=csv

// infraProvider provides the infrastructure primitives this repository
type infraProvider interface {
	// CsvNewReader returns a new csv.Reader that reads from r.
	CsvNewReader(r io.Reader) *csv.Reader
	// CsvReadAll reads all the remaining records from r.
	CsvReadAll(r gocsv.Reader) ([][]string, error)
	// TimeTruncateToDay strips the time-of-day component from t.
	TimeTruncateToDay(t time.Time) time.Time
	// TimeInRange reports whether d's date falls within [start, end] inclusive.
	TimeInRange(d, start, end time.Time) bool
}

// Repository type of csv-backed repository.
type Repository struct {
	infra infraProvider
}

// NewRepository instantiates the csv repository with the given
// infrastructure dependency.
func NewRepository(infra infraProvider) *Repository {
	return &Repository{infra: infra}
}
