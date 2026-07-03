package bank

import "time"

type BankStatementParams struct {
	BankName string
	UserID   int64
	Path     string
	Start    time.Time // YYYY-MM-DD
	End      time.Time // YYYY-MM-DD
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
