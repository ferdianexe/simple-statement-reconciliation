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
