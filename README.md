# simple-statement-reconciliation
This is a simple transaction reconciliation service built with Go.
## Features
- Reconcile internal system transactions against one or more bank statement CSV files
- Detect matched, discrepant, and unmatched transactions within a date range
- Group unmatched bank records by bank
- Run as a one-off CLI batch job or as an HTTP service
- Reconcile from server-side file paths, or directly from uploaded CSV content
- Output as plain text or JSON
## Notes
- A system transaction and a bank record are matched on date + type first; amount is only compared afterward, since discrepancies only ever occur in amount.
- If more than one bank record could match a transaction (same date, same type, same amount), the closest amount is picked and ties are broken by file order.
- A bank record's amount sign indicates direction: negative = DEBIT, positive = CREDIT.
- Matching is one-to-one. A transaction and a bank record can each be used in at most one match.
- Date range filtering is inclusive on both ends.
- CSV input files must include a header row.
## How to run
1. Install Go 1.21+
2. Clone the repository
3. Run `go mod tidy` to fetch dependencies
4. Run `make run` (CLI) or `make serve` (HTTP service)
## How to use
### Run as a CLI
```
make run SYS=testdata/system_transactions.csv BANKS="BCA:testdata/bank_bca.csv,BNI:testdata/bank_bni.csv" START=2024-01-01 END=2024-01-31
```
Response sample:
```
Total processed:   11
Total matched:     4
Total unmatched:   3
Total discrepancy: 5000.00

Unmatched system transactions:
  - TRX004 | 110000.00 | DEBIT | 2024-01-10T09:00:00Z

Unmatched bank statement records:
  BCA:
    - BCA-99190 | -50000.00 | 2024-01-15
  BNI:
    - BNI-4480 | 75000.00 | 2024-01-20
```
Add `-json` to `make run-json` for machine-readable output instead.
### Run as an HTTP service
```
POST /reconcile
Content-Type: application/json
{
    "sys_path": "testdata/system_transactions.csv",
    "banks": [
        {"name": "BCA", "path": "testdata/bank_bca.csv"},
        {"name": "BNI", "path": "testdata/bank_bni.csv"}
    ],
    "start": "2024-01-01",
    "end": "2024-01-31"
}
```
Response sample:
```
{
    "TotalProcessed": 11,
    "TotalMatched": 4,
    "TotalUnmatched": 3,
    "TotalDiscrepancy": 5000,
    "Matched": [
        {
            "System": {"TrxID": "TRX005", "Amount": 220000, "Type": "CREDIT", "TransactionTime": "2024-01-12T15:00:00Z"},
            "Bank": {"UniqueID": "BNI-4471", "Amount": 220000, "Date": "2024-01-12T00:00:00Z", "Bank": "BNI"},
            "Discrepancy": 0
        }
    ],
    "UnmatchedSystem": [
        {"TrxID": "TRX004", "Amount": 110000, "Type": "DEBIT", "TransactionTime": "2024-01-10T09:00:00Z"}
    ],
    "UnmatchedBank": [
        {"Bank": "BCA", "Records": [{"UniqueID": "BCA-99190", "Amount": -50000, "Date": "2024-01-15T00:00:00Z", "Bank": "BCA"}]},
        {"Bank": "BNI", "Records": [{"UniqueID": "BNI-4480", "Amount": 75000, "Date": "2024-01-20T00:00:00Z", "Bank": "BNI"}]}
    ]
}
```
### Import via CSV upload
```
POST /reconcile/import
Content-Type: multipart/form-data

start=2024-01-01
end=2024-01-31
system_transactions=<system_transactions.csv file>
BCA=<bank_bca.csv file>
BNI=<bank_bni.csv file>
```
Unlike `/reconcile`, this endpoint takes CSV file content directly instead of server-side file paths. The `system_transactions` field name is reserved for csv system transaction; while csv bank statement files are optional. but at least one of them must be provided. has identical response format to `/reconcile`.
Supported bank names are:
- BCA
- BNI
- Mandiri
- BRI

Example with curl:
```
curl -X POST localhost:8080/reconcile/import \
  -F start=2024-01-01 \
  -F end=2024-01-31 \
  -F system_transactions=@testdata/system_transactions.csv \
  -F BCA=@testdata/bank_bca.csv \
  -F BNI=@testdata/bank_bni.csv
```
### Health check
```
GET /healthz
```
Response sample:
```
ok
```
