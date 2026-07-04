package infrastructure

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
)

// timeManager holds methods for golang Time.
type timeManager interface {
	// TruncateToDay strips the time-of-day component. Useful when one side
	// of a comparison carries a full timestamp and the other only carries
	// a date, since both need to be compared at day granularity.
	TruncateToDay(t time.Time) time.Time

	// Now returns the current local time.
	Now() time.Time

	// InRange reports whether d's date falls within [start, end] inclusive.
	// start and end are expected to already be day-truncated by the caller.
	InRange(d, start, end time.Time) bool
}

// csvManager holds methods for csv
type csvManager interface {
	// NewReader returns a new Reader that reads from r.
	NewReader(r io.Reader) *csv.Reader

	// Read reads one record (a slice of fields) from r.
	Read(r gocsv.Reader) (record []string, err error)

	// ReadAll reads all the remaining records from r.
	// Each record is a slice of fields.
	// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
	// defined to read until EOF, it does not treat end of file as an error to be
	// reported.
	ReadAll(r gocsv.Reader) ([][]string, error)
}

// NewServiceParam is the set of dependencies required to construct a
// new infrastructure Service.
type NewServiceParam struct {
	Csv  csvManager
	Time timeManager
}

// Service is an internal implementation of infrastructure service.
// It wraps methods for provided infras.
type Service struct {
	csvManager  csvManager
	timeManager timeManager
}

// NewService returns an instance of the infrastructure Service.
func NewService(param NewServiceParam) *Service {
	return &Service{
		csvManager:  param.Csv,
		timeManager: param.Time,
	}
}

// CsvNewReader returns a new csv.Reader that reads from r.
func (svc *Service) CsvNewReader(r io.Reader) *csv.Reader {
	return svc.csvManager.NewReader(r)
}

// CsvRead reads one record (a slice of fields) from r.
func (svc *Service) CsvRead(r gocsv.Reader) ([]string, error) {
	return svc.csvManager.Read(r)
}

// CsvReadAll reads all the remaining records from r.
func (svc *Service) CsvReadAll(r gocsv.Reader) ([][]string, error) {
	return svc.csvManager.ReadAll(r)
}

// TimeTruncateToDay strips the time-of-day component from t.
func (svc *Service) TimeTruncateToDay(t time.Time) time.Time {
	return svc.timeManager.TruncateToDay(t)
}

// TimeInRange reports whether d's date falls within [start, end] inclusive.
func (svc *Service) TimeInRange(d, start, end time.Time) bool {
	return svc.timeManager.InRange(d, start, end)
}

// GetTimeNow get time now.
//
// It returns time.Time when successful.
// Otherwise, empty time.Time will be returned.
func (svc *Service) GetTimeNow() time.Time {
	return svc.timeManager.Now()
}
