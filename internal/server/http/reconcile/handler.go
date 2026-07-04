package reconcile

import (
	"context"
	"encoding/csv"
	"io"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=reconcile

// reconcileUsecaseManager provides the required action usecase methods.
type reconcileUsecaseManager interface {
	// Reconcile performs reconciliation process.
	Reconcile(ctx context.Context, request reconcile.ReconcileRequest) (reconcile.ReconcileSummary, error)
	// ReconcileFromRecords performs reconciliation on already-parsed and
	// already-filtered records, used by the CSV-upload endpoint.
	ReconcileFromRecords(ctx context.Context, sysTrx []reconcile.SystemTransaction, bankStmts []reconcile.BankStatement, start, end time.Time) (reconcile.ReconcileSummary, error)
}

// infraProvider provides the infrastructure primitives this handler
// needs for reading uploaded CSV content and applying date-range
// filtering, injected so both can be mocked in tests rather than this
// layer calling encoding/csv or gotime directly.
type infraProvider interface {
	// CsvNewReader returns a new csv.Reader that reads from r.
	CsvNewReader(r io.Reader) *csv.Reader
	// CsvReadAll reads all the remaining records from r.
	CsvReadAll(r gocsv.Reader) ([][]string, error)
	// TimeInRange reports whether d's date falls within [start, end] inclusive.
	TimeInRange(d, start, end time.Time) bool
}

// Handler type of reconcile handler.
type Handler struct {
	reconcile reconcileUsecaseManager
	infra     infraProvider
}

// NewHandler instantiates reconcile handler.
func NewHandler(reconcile reconcileUsecaseManager, infra infraProvider) *Handler {
	return &Handler{
		reconcile: reconcile,
		infra:     infra,
	}
}
