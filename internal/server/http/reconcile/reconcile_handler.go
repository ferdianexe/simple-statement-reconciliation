package reconcile

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
)

const dateLayout = "2006-01-02"

func (h *Handler) HandleReconcile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}

	var req ReconcileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	start, err := time.Parse(dateLayout, req.Start)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start date, expected YYYY-MM-DD: "+err.Error())
		return
	}
	end, err := time.Parse(dateLayout, req.End)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end date, expected YYYY-MM-DD: "+err.Error())
		return
	}

	banks := make([]reconcile.BankFileRequest, 0, len(req.Banks))
	for _, b := range req.Banks {
		banks = append(banks, reconcile.BankFileRequest{Name: b.Name, Path: b.Path})
	}

	summary, err := h.reconcile.Reconcile(r.Context(), reconcile.ReconcileRequest{
		SysPath: req.SysPath,
		Banks:   banks,
		Start:   start,
		End:     end,
	})
	if err != nil {
		// Input/file problems are the client's fault here (bad paths,
		// malformed CSV), so 422 rather than 500.
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	matched := make([]MatchedPair, 0, len(summary.Matched))
	for _, m := range summary.Matched {
		matched = append(matched, MatchedPair{
			System: SystemTransaction{
				TrxID:           m.System.TrxID,
				Type:            TrxType(m.System.Type),
				Amount:          m.System.Amount,
				TransactionTime: m.System.TransactionTime,
			},
			Bank: BankStatement{
				UniqueID: m.Bank.UniqueID,
				Amount:   m.Bank.Amount,
				Date:     m.Bank.Date,
				Bank:     m.Bank.Bank,
			},
			Discrepancy: m.Discrepancy,
		})
	}

	unmatchedSystem := make([]SystemTransaction, 0, len(summary.UnmatchedSystem))
	for _, u := range summary.UnmatchedSystem {
		unmatchedSystem = append(unmatchedSystem, SystemTransaction{
			TrxID:           u.TrxID,
			Type:            TrxType(u.Type),
			Amount:          u.Amount,
			TransactionTime: u.TransactionTime,
		})
	}

	unmatchedBank := make([]UnmatchedByBank, 0, len(summary.UnmatchedBank))
	for _, u := range summary.UnmatchedBank {
		recordBankStmt := make([]BankStatement, 0, len(u.Records))
		for _, r := range u.Records {
			recordBankStmt = append(recordBankStmt, BankStatement{
				UniqueID: r.UniqueID,
				Amount:   r.Amount,
				Date:     r.Date,
				Bank:     r.Bank,
			})
		}
		unmatchedBank = append(unmatchedBank, UnmatchedByBank{
			Bank:    u.Bank,
			Records: recordBankStmt,
		})
	}

	srvResult := ReconcileSummary{
		TotalProcessed:   summary.TotalProcessed,
		TotalMatched:     summary.TotalMatched,
		TotalUnmatched:   summary.TotalUnmatched,
		TotalDiscrepancy: summary.TotalDiscrepancy,
		Matched:          matched,
		UnmatchedSystem:  unmatchedSystem,
		UnmatchedBank:    unmatchedBank,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(srvResult); err != nil {
		log.Printf("encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
