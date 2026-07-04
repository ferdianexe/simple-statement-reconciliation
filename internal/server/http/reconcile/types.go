package reconcile

import "time"

// TrxType represents the direction of a transaction.
// moves to constant package
type TrxType string

const (
	Debit  TrxType = "DEBIT"
	Credit TrxType = "CREDIT"
)

// bankFileRequest mirrors service.BankFile for JSON (de)serialization.
type BankFileRequest struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// reconcileRequest is the expected POST /reconcile body. File paths are
// used rather than uploaded file contents to keep this take-home-sized;
// see README for the multipart-upload extension this would need in a
// real deployment.
type ReconcileRequest struct {
	SysPath string            `json:"sys_path"`
	Banks   []BankFileRequest `json:"banks"`
	Start   string            `json:"start"` // YYYY-MM-DD
	End     string            `json:"end"`   // YYYY-MM-DD
}

type errorResponse struct {
	Error string `json:"error"`
}

// ReconcileSummary is the full output of a reconciliation run.
type ReconcileSummary struct {
	TotalProcessed       int
	TotalMatched         int
	TotalUnmatched       int
	TotalUnmatchedAmount int
	TotalUnmatchedBank   int
	TotalUnmatchedSystem int
	TotalDiscrepancy     float64

	Matched         []MatchedPair
	AmountMismatch  []MatchedPair
	UnmatchedSystem []SystemTransaction
	UnmatchedBank   []UnmatchedByBank
}

// MatchedPair is a system transaction successfully paired with a bank
// statement record. Discrepancy is the absolute difference in amount
// between the two sides - per the problem statement, discrepancies only
// occur in amount, so date/type are treated as reliable matching keys
// and amount is compared only *after* a match is found.
type MatchedPair struct {
	System      SystemTransaction
	Bank        BankStatement
	Discrepancy float64
}

// BankStatement is a record from an external bank statement file.
// Amount can be negative (debit) or positive (credit) depending on the
// bank's own convention; Bank identifies which source file it came from
// since the service can reconcile against multiple banks at once.
type BankStatement struct {
	UniqueID string
	Amount   float64
	Date     time.Time
	Bank     string
}

// SystemTransaction is a record from Amartha's internal system.
type SystemTransaction struct {
	TrxID           string
	Amount          float64 // always positive; direction is carried by Type
	Type            TrxType
	TransactionTime time.Time
}

// UnmatchedByBank groups unmatched bank statement records by their
// originating bank, as required by the output spec.
type UnmatchedByBank struct {
	Bank    string
	Records []BankStatement
}
