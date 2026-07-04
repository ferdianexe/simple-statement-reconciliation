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

// Reconcile loads the system transaction history and bank statement
// history required by request, then matches them via doReconcile.
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

// ReconcileFromRecords reconciles already-parsed system transactions and
// bank statement records directly, filtering out records outside the start/end range.
func (uc *Usecase) ReconcileFromRecords(ctx context.Context, sysTrx []SystemTransaction, bankStmts []BankStatement, start, end time.Time) (ReconcileSummary, error) {
	return uc.doReconcile(sysTrx, bankStmts, start, end), nil
}

// key builds the (date, type) bucket key. Both sides are normalized to a
// day-granularity date, since system transactions carry a full
// timestamp but bank statements only carry a date.
func key(date time.Time, t TrxType) string {
	return date.Format("2006-01-02") + "|" + string(t)
}

// doReconcile matches system transactions against bank statement records
// within [start, end] (inclusive) and returns a full summary.
func (uc *Usecase) doReconcile(sysTrx []SystemTransaction, bankStmts []BankStatement, start, end time.Time) ReconcileSummary {
	start = uc.infra.TimeTruncateToDay(start)
	end = uc.infra.TimeTruncateToDay(end)

	// Filter system transactions to the start/end range.
	filteredSys := make([]SystemTransaction, 0, len(sysTrx))
	for _, t := range sysTrx {
		if uc.infra.TimeInRange(t.TransactionTime, start, end) {
			filteredSys = append(filteredSys, t)
		}
	}

	// sort first to make sure its ordered.
	sort.Slice(filteredSys, func(i, j int) bool {
		return filteredSys[i].TransactionTime.Before(filteredSys[j].TransactionTime)
	})

	// bank statement need to be filtered to the start/end range. for safeguard purpose
	// we dont need to sort order this, since reconcile will search a while based on key (time, inferred type)
	// next we will search the closest different amount bank statement with current system transaction
	buckets := make(map[string][]*bankEntry)
	for _, b := range bankStmts {
		if !uc.infra.TimeInRange(b.Date, start, end) {
			continue
		}
		k := key(b.Date, b.InferredType())
		buckets[k] = append(buckets[k], &bankEntry{stmt: b})
	}

	var matched []MatchedPair
	var amountMismatch []MatchedPair
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
			// only compare absolute amount and select the closest amount
			diff := math.Abs(s.Amount - c.stmt.AbsAmount())
			if diff < bestDiff {
				bestDiff = diff
				best = i
			} else if diff == 0 {
				// absolute match
				bestDiff = diff
				best = i
				break
			}
		}

		if best == -1 {
			unmatchedSys = append(unmatchedSys, s)
			continue
		}

		if bestDiff == 0 {
			// absolute match
			candidates[best].matched = true
			matched = append(matched, MatchedPair{
				System: s,
				Bank:   candidates[best].stmt,
			})
		} else {
			// amount mismatch
			candidates[best].matched = true
			disc := bestDiff
			totalDiscrepancy += disc
			amountMismatch = append(amountMismatch, MatchedPair{
				System:      s,
				Bank:        candidates[best].stmt,
				Discrepancy: disc,
			})
		}

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
		TotalProcessed:       len(filteredSys) + totalBankCount,
		TotalMatched:         len(matched),
		TotalUnmatched:       len(unmatchedSys) + totalUnmatchedBank + len(amountMismatch),
		TotalUnmatchedAmount: len(amountMismatch),
		TotalUnmatchedBank:   totalUnmatchedBank,
		TotalUnmatchedSystem: len(unmatchedSys),
		TotalDiscrepancy:     totalDiscrepancy,
		Matched:              matched,
		AmountMismatch:       amountMismatch,
		UnmatchedSystem:      unmatchedSys,
		UnmatchedBank:        unmatchedBank,
	}
}
