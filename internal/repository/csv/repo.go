package csv

// Repository type of DS model repository that used as collection of DS model client.
type Repository struct {
}

// NewRepository Package parser reads the two input CSV formats into domain models.
//
// Expected system transaction CSV columns (header row required):
//
//	trx_id,amount,type,transaction_time
//	TRX001,110000,DEBIT,2024-01-08T10:00:00Z
//
// Expected bank statement CSV columns (header row required):
//
//	unique_identifier,amount,date
//	BCA-99182,-110000,2024-01-08
//
// Both parsers stream the file row by row rather than loading it into
// memory as a whole string, so behaviour degrades gracefully as input
// files grow large.
func NewRepository() *Repository {
	return &Repository{}
}
