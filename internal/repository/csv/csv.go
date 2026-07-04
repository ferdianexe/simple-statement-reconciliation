package csv

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	systemTimeLayout = time.RFC3339
	bankDateLayout   = "2006-01-02"
)

// ParseSystemTransactions reads the internal system transaction CSV
func (svc *Repository) ParseSystemTransactions(ctx context.Context, path string, start, end time.Time) ([]Transaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open system transactions file: %w", err)
	}
	defer f.Close()

	start = svc.infra.TimeTruncateToDay(start)
	end = svc.infra.TimeTruncateToDay(end)

	reader := svc.infra.CsvNewReader(f)
	records, err := svc.infra.CsvReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	out := make([]Transaction, 0, len(records)-1)
	for i, rec := range records[1:] { // records[0] is the header
		line := i + 2
		if len(rec) < 4 {
			return nil, fmt.Errorf("line %d: expected 4 columns, got %d", line, len(rec))
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid amount %q: %w", line, rec[1], err)
		}
		trxType := TrxType(strings.ToUpper(strings.TrimSpace(rec[2])))
		if trxType != Debit && trxType != Credit {
			return nil, fmt.Errorf("line %d: invalid type %q", line, rec[2])
		}
		ts, err := time.Parse(systemTimeLayout, strings.TrimSpace(rec[3]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid transaction_time %q: %w", line, rec[3], err)
		}

		if !svc.infra.TimeInRange(ts, start, end) {
			continue
		}

		out = append(out, Transaction{
			TrxID:           strings.TrimSpace(rec[0]),
			Amount:          amount,
			Type:            trxType,
			TransactionTime: ts,
		})
	}
	return out, nil
}

// ParseBankStatement reads one bank statement CSV file
func (svc *Repository) ParseBankStatement(ctx context.Context, path string, bankName string, start, end time.Time) ([]BankStatement, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open bank statement file %s: %w", path, err)
	}
	defer f.Close()

	start = svc.infra.TimeTruncateToDay(start)
	end = svc.infra.TimeTruncateToDay(end)

	reader := svc.infra.CsvNewReader(f)
	records, err := svc.infra.CsvReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%s: read csv: %w", bankName, err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	out := make([]BankStatement, 0, len(records)-1)
	for i, rec := range records[1:] { // records[0] is the header
		line := i + 2
		if len(rec) < 3 {
			return nil, fmt.Errorf("%s line %d: expected 3 columns, got %d", bankName, line, len(rec))
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: invalid amount %q: %w", bankName, line, rec[1], err)
		}
		date, err := time.Parse(bankDateLayout, strings.TrimSpace(rec[2]))
		if err != nil {
			return nil, fmt.Errorf("%s line %d: invalid date %q: %w", bankName, line, rec[2], err)
		}

		if !svc.infra.TimeInRange(date, start, end) {
			continue
		}

		out = append(out, BankStatement{
			UniqueID: strings.TrimSpace(rec[0]),
			Amount:   amount,
			Date:     date,
			Bank:     bankName,
		})
	}
	return out, nil
}
