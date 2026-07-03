package transaction

import "time"

type TransactionHistoryParams struct {
	UserID    int64
	SysPath   string
	StartDate time.Time
	EndDate   time.Time
}

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
