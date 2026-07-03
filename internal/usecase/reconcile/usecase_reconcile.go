package reconcile

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/bank"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/transaction"
)

func (uc *Usecase) Reconcile(ctx context.Context, request ReconcileRequest) (ReconcileSummary, error) {
	banks := make([]bank.BankStatementParams, 0, len(request.Banks))
	for _, b := range request.Banks {
		banks = append(banks, bank.BankStatementParams{
			BankName: b.Name,
			Path:     b.Path,
			Start:    request.Start,
			End:      request.End,
		})
	}

	userTrx, err := uc.transaction.GetUserTransactionHistory(ctx, transaction.TransactionHistoryParams{SysPath: request.SysPath})
	if err != nil {
		return ReconcileSummary{}, fmt.Errorf("parsing user transactions: %w", err)
	}

	ucUserTrx := make([]SystemTransaction, 0, len(userTrx))
	for _, t := range userTrx {
		ucUserTrx = append(ucUserTrx, SystemTransaction{
			TransactionTime: t.TransactionTime,
			Type:            TrxType(t.Type),
			Amount:          t.Amount,
			TrxID:           t.TrxID,
		})
	}

	bankStatements, err := uc.bank.GetBankStatementHistory(ctx, banks)
	if err != nil {
		return ReconcileSummary{}, fmt.Errorf("parsing bank statements: %w", err)
	}

	ucBankStatement := make([]BankStatement, 0, len(bankStatements))
	for _, b := range bankStatements {
		ucBankStatement = append(ucBankStatement, BankStatement{
			Date:     b.Date,
			Amount:   b.Amount,
			UniqueID: b.UniqueID,
			Bank:     b.Bank,
		})
	}

	return uc.doReconcile(ucUserTrx, ucBankStatement, request.Start, request.End), nil
}

// key builds the (date, type) bucket key. Both sides are normalized to a
// day-granularity date, since system transactions carry a full
// timestamp but bank statements only carry a date.
func key(date time.Time, t TrxType) string {
	return date.Format("2006-01-02") + "|" + string(t)
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func inRange(d, start, end time.Time) bool {
	d = truncateToDay(d)
	return !d.Before(start) && !d.After(end)
}

// doReconcile matches system transactions against bank statement records
// within [start, end] (inclusive) and returns a full summary.
func (uc *Usecase) doReconcile(sysTrx []SystemTransaction, bankStmts []BankStatement, start, end time.Time) ReconcileSummary {
	start = truncateToDay(start)
	end = truncateToDay(end)

	filteredSys := make([]SystemTransaction, 0, len(sysTrx))
	for _, t := range sysTrx {
		if inRange(t.TransactionTime, start, end) {
			filteredSys = append(filteredSys, t)
		}
	}
	// Process in a stable, deterministic order so re-running the same
	// input always produces the same matches.
	sort.Slice(filteredSys, func(i, j int) bool {
		return filteredSys[i].TransactionTime.Before(filteredSys[j].TransactionTime)
	})

	// Index bank statements into buckets keyed by (date, inferred type).
	buckets := make(map[string][]*bankEntry)
	for _, b := range bankStmts {
		if !inRange(b.Date, start, end) {
			continue
		}
		k := key(b.Date, b.InferredType())
		buckets[k] = append(buckets[k], &bankEntry{stmt: b})
	}

	var matched []MatchedPair
	var unmatchedSys []SystemTransaction
	var totalDiscrepancy float64

	for _, s := range filteredSys {
		k := key(s.TransactionTime, s.Type)
		candidates := buckets[k]

		best := -1
		bestDiff := math.MaxFloat64
		for i, c := range candidates {
			if c.matched {
				continue
			}
			diff := math.Abs(s.Amount - c.stmt.AbsAmount())
			if diff < bestDiff {
				bestDiff = diff
				best = i
			}
		}

		if best == -1 {
			unmatchedSys = append(unmatchedSys, s)
			continue
		}

		candidates[best].matched = true
		disc := bestDiff
		totalDiscrepancy += disc
		matched = append(matched, MatchedPair{
			System:      s,
			Bank:        candidates[best].stmt,
			Discrepancy: disc,
		})
	}

	// Collect remaining unmatched bank records, grouped by bank.
	unmatchedByBank := map[string][]BankStatement{}
	bankOrder := make([]string, 0)
	totalBankCount := 0
	for _, entries := range buckets {
		for _, e := range entries {
			totalBankCount++
			if e.matched {
				continue
			}
			if _, ok := unmatchedByBank[e.stmt.Bank]; !ok {
				bankOrder = append(bankOrder, e.stmt.Bank)
			}
			unmatchedByBank[e.stmt.Bank] = append(unmatchedByBank[e.stmt.Bank], e.stmt)
		}
	}
	sort.Strings(bankOrder)

	var unmatchedBank []UnmatchedByBank
	for _, bank := range bankOrder {
		records := unmatchedByBank[bank]
		sort.Slice(records, func(i, j int) bool { return records[i].Date.Before(records[j].Date) })
		unmatchedBank = append(unmatchedBank, UnmatchedByBank{Bank: bank, Records: records})
	}

	totalUnmatchedBank := 0
	for _, g := range unmatchedBank {
		totalUnmatchedBank += len(g.Records)
	}

	return ReconcileSummary{
		TotalProcessed:   len(filteredSys) + totalBankCount,
		TotalMatched:     len(matched),
		TotalUnmatched:   len(unmatchedSys) + totalUnmatchedBank,
		TotalDiscrepancy: totalDiscrepancy,
		Matched:          matched,
		UnmatchedSystem:  unmatchedSys,
		UnmatchedBank:    unmatchedBank,
	}
}
