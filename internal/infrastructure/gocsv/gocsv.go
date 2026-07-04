package gocsv

import (
	// golang package
	"encoding/csv"
	"io"
)

//go:generate mockgen -source=./gocsv.go -destination=./gocsv_mock.go -package=gocsv

// Default represents the default instance of goCsv.
var Default = goCsv{}

// goCsv represents an empty struct wrapping encoding/csv.
type goCsv struct{}

// Reader represents the method interface for reading CSV records. Since
// *csv.Reader (returned by NewReader) already implements Read and
// ReadAll, it satisfies this interface directly - callers can pass
// either a real *csv.Reader or a mock wherever Reader is expected.
type Reader interface {
	// Read reads one record (a slice of fields) from r.
	Read() (record []string, err error)

	// ReadAll reads all the remaining records from r.
	// Each record is a slice of fields.
	// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
	// defined to read until EOF, it does not treat end of file as an error to be
	// reported.
	ReadAll() (records [][]string, err error)
}

// NewReader returns a new Reader that reads from r.
func (g goCsv) NewReader(r io.Reader) *csv.Reader {
	return csv.NewReader(r)
}

// Read reads one record (a slice of fields) from r. A returned error of
// io.EOF signals the end of input, same as the underlying csv.Reader.
func (g goCsv) Read(r Reader) (record []string, err error) {
	return r.Read()
}

// ReadAll reads all the remaining records from r.
// Each record is a slice of fields.
// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
// defined to read until EOF, it does not treat end of file as an error to be
// reported.
func (g goCsv) ReadAll(r Reader) ([][]string, error) {
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}
