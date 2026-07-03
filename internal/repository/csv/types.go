package csv

import "time"

// TrxType represents the direction of a transaction.
type TrxType string

const (
	Debit  TrxType = "DEBIT"
	Credit TrxType = "CREDIT"
)

// Transaction is a record from Amartha's internal system.
type Transaction struct {
	TrxID           string
	Amount          float64 // always positive; direction is carried by Type
	Type            TrxType
	TransactionTime time.Time
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
	System      Transaction
	Bank        BankStatement
	Discrepancy float64
}

// UnmatchedByBank groups unmatched bank statement records by their
// originating bank, as required by the output spec.
type UnmatchedByBank struct {
	Bank    string
	Records []BankStatement
}

// Summary is the full output of a reconciliation run.
type Summary struct {
	TotalProcessed   int
	TotalMatched     int
	TotalUnmatched   int
	TotalDiscrepancy float64

	Matched         []MatchedPair
	UnmatchedSystem []Transaction
	UnmatchedBank   []UnmatchedByBank
}
