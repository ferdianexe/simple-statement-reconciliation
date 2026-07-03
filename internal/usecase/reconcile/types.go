package reconcile

import (
	"time"
)

// TrxType represents the direction of a transaction.
// moves to constant package
type TrxType string

const (
	Debit  TrxType = "DEBIT"
	Credit TrxType = "CREDIT"
)

// reconcileRequest is the expected POST /reconcile body. File paths are
// used rather than uploaded file contents to keep this take-home-sized;
// see README for the multipart-upload extension this would need in a
// real deployment.
type ReconcileRequest struct {
	SysPath string
	Banks   []BankFileRequest
	Start   time.Time // YYYY-MM-DD
	End     time.Time // YYYY-MM-DD
}

// BankFileRequest
type BankFileRequest struct {
	Name string
	Path string
}

// bankEntry wraps a bank statement record with matched-state tracked
// during the algorithm run.
type bankEntry struct {
	stmt    BankStatement
	matched bool
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

// AbsAmount returns the unsigned amount of the bank record.
func (b BankStatement) AbsAmount() float64 {
	if b.Amount < 0 {
		return -b.Amount
	}
	return b.Amount
}

// InferredType derives DEBIT/CREDIT from the sign of the bank amount.
// Assumption: negative amount = money leaving the account = DEBIT,
// positive amount = money entering the account = CREDIT. This mirrors
// the system Transaction.Type semantics so the two sources can be
// matched on a common (date, type) key.
func (b BankStatement) InferredType() TrxType {
	if b.Amount < 0 {
		return Debit
	}
	return Credit
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

// ReconcileSummary is the full output of a reconciliation run.
type ReconcileSummary struct {
	TotalProcessed   int
	TotalMatched     int
	TotalUnmatched   int
	TotalDiscrepancy float64

	Matched         []MatchedPair
	UnmatchedSystem []SystemTransaction
	UnmatchedBank   []UnmatchedByBank
}
