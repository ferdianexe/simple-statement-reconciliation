package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	systemTimeLayout = time.RFC3339
	bankDateLayout   = "2006-01-02"
)

// ParseSystemTransactions reads the internal system transaction CSV file.
func (svc *Repository) ParseSystemTransactions(ctx context.Context, path string) ([]Transaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open system transactions file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	if _, err := r.Read(); err != nil { // header
		return nil, fmt.Errorf("read header: %w", err)
	}

	var out []Transaction
	line := 1
	for {
		line++
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
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

		out = append(out, Transaction{
			TrxID:           strings.TrimSpace(rec[0]),
			Amount:          amount,
			Type:            trxType,
			TransactionTime: ts,
		})
	}
	return out, nil
}

// ParseBankStatement reads one bank statement CSV file. bankName tags every
// record so unmatched results can later be grouped per bank, since the
// service supports reconciling against several banks in a single run.
func (svc *Repository) ParseBankStatement(ctx context.Context, path string, bankName string) ([]BankStatement, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open bank statement file %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	if _, err := r.Read(); err != nil { // header
		return nil, fmt.Errorf("read header: %w", err)
	}

	var out []BankStatement
	line := 1
	for {
		line++
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", bankName, line, err)
		}
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

		out = append(out, BankStatement{
			UniqueID: strings.TrimSpace(rec[0]),
			Amount:   amount,
			Date:     date,
			Bank:     bankName,
		})
	}
	return out, nil
}
