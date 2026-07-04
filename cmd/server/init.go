package main

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/server/http/reconcile"
)

//go:generate mockgen -source=./init.go -destination=./init_mock.go -package=main

// HTTPAppHandlers contains list of http and nsq handler to support main app.
type HTTPAppHandlers struct {
	HTTP HTTPHandlers
}

// HTTPHandlers contains a list of http handler to support main app.
type HTTPHandlers struct {
	Reconcile *reconcile.Handler
}

// infraProvider is the interface that wraps the infrastructure methods needed by handler.
type infraProvider interface {
	// CsvNewReader returns a new csv.Reader that reads from r.
	CsvNewReader(r io.Reader) *csv.Reader
	// CsvReadAll reads all the remaining records from r.
	CsvReadAll(r gocsv.Reader) ([][]string, error)
	// TimeInRange reports whether d's date falls within [start, end] inclusive.
	TimeInRange(d, start, end time.Time) bool
	// TimeTruncateToDay strips the time-of-day component from t.
	TimeTruncateToDay(t time.Time) time.Time
}

// NewHTTPAppHandlers initialize handlers with each use case according to the requirement for each handlers.
// It accepts ucs, writer, infra, featureFlag as parameters.
// It returns non nil pointer HTTPAppHandlers.
func NewHTTPAppHandlers(ucs *Usecases, infra infraProvider) *HTTPAppHandlers {
	h := &HTTPAppHandlers{
		HTTP: HTTPHandlers{
			Reconcile: reconcile.NewHandler(ucs.reconcile, infra),
		},
	}

	return h
}
