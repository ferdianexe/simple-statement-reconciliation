package reconcile

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
)

const (
	dateLayout = "2006-01-02"
	// importMaxMemory mirrors net/http's own default for
	// ParseMultipartForm: parts under this size are kept in memory,
	// anything larger spills to temp files on disk.
	importMaxMemory = 32 << 20 // 32MB

	// systemTransactionsField is the reserved multipart field name for
	// the system transaction CSV file.
	systemTransactionsField = "system_transactions"
)

var (
	// allowedBanks maps every accepted bank identifier (matched
	// case-insensitively) to its canonical display name
	allowedBanks = map[string]string{
		"BCA":     "BCA",
		"BNI":     "BNI",
		"BRI":     "BRI",
		"MANDIRI": "Mandiri",
	}
)

// HandleReconcile handles POST /reconcile. It expects a JSON body with
// SysPath, Banks (file paths), Start, and End (YYYY-MM-DD), and returns
// the reconciliation summary as JSON.
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

	writeReconcileSummary(w, summary)
}

// HandleReconcileImport handles POST /reconcile/import. It expects a
// multipart form with "start"/"end" (YYYY-MM-DD) and CSV file fields
// (system_transactions plus one per bank), and returns the same
// reconciliation summary shape as HandleReconcile.
func (h *Handler) HandleReconcileImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}

	if err := r.ParseMultipartForm(importMaxMemory); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		return
	}

	start, err := time.Parse(dateLayout, r.FormValue("start"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start date, expected YYYY-MM-DD: "+err.Error())
		return
	}
	end, err := time.Parse(dateLayout, r.FormValue("end"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end date, expected YYYY-MM-DD: "+err.Error())
		return
	}

	sysTrx, err := h.parseSystemTransactionsMultipart(r, start, end)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, systemTransactionsField+": "+err.Error())
		return
	}

	bankStmts, err := h.parseBankStatementsMultipart(r, start, end)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	summary, err := h.reconcile.ReconcileFromRecords(r.Context(), sysTrx, bankStmts, start, end)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeReconcileSummary(w, summary)
}

// parseSystemTransactionsMultipart reads and date-filters the
// "system_transactions" file field, if present. A missing field returns
// a nil slice and no error - the request simply has no system-side data.
func (h *Handler) parseSystemTransactionsMultipart(r *http.Request, start, end time.Time) ([]reconcile.SystemTransaction, error) {
	file, _, err := r.FormFile(systemTransactionsField)
	if err == http.ErrMissingFile {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := h.infra.CsvNewReader(file)
	records, err := h.infra.CsvReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	out := make([]reconcile.SystemTransaction, 0, len(records)-1)
	for i, rec := range records[1:] { // records[0] is the header
		line := i + 2
		if len(rec) < 4 {
			return nil, fmt.Errorf("line %d: expected 4 columns, got %d", line, len(rec))
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid amount %q: %w", line, rec[1], err)
		}
		trxType := reconcile.TrxType(strings.ToUpper(strings.TrimSpace(rec[2])))
		if trxType != reconcile.Debit && trxType != reconcile.Credit {
			return nil, fmt.Errorf("line %d: invalid type %q", line, rec[2])
		}
		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(rec[3]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid transaction_time %q: %w", line, rec[3], err)
		}

		if !h.infra.TimeInRange(ts, start, end) {
			continue
		}

		out = append(out, reconcile.SystemTransaction{
			TrxID:           strings.TrimSpace(rec[0]),
			Amount:          amount,
			Type:            trxType,
			TransactionTime: ts,
		})
	}
	return out, nil
}

// parseBankStatementsMultipart reads every uploaded bank statement file
// field name against allowedBanks, and returns the combined, date-
// filtered records. No bank files at all is not an error - it just
// means there's nothing to reconcile on the bank side.
func (h *Handler) parseBankStatementsMultipart(r *http.Request, start, end time.Time) ([]reconcile.BankStatement, error) {
	if r.MultipartForm == nil {
		return nil, nil
	}

	var out []reconcile.BankStatement
	for field, headers := range r.MultipartForm.File {
		if field == systemTransactionsField || len(headers) == 0 {
			continue
		}

		bankName, ok := allowedBanks[strings.ToUpper(field)]
		if !ok {
			return nil, fmt.Errorf("%s: unsupported bank, allowed: %s", field, allowedBankNames())
		}

		file, err := headers[0].Open()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field, err)
		}

		stmts, err := h.parseBankStatementFile(file, bankName, start, end)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field, err)
		}
		out = append(out, stmts...)
	}
	return out, nil
}

// parseBankStatementFile parses a single bank statement CSV, tagging
// every record with bankName and filtering out rows outside [start, end].
func (h *Handler) parseBankStatementFile(file multipart.File, bankName string, start, end time.Time) ([]reconcile.BankStatement, error) {
	reader := h.infra.CsvNewReader(file)
	records, err := h.infra.CsvReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	out := make([]reconcile.BankStatement, 0, len(records)-1)
	for i, rec := range records[1:] { // records[0] is the header
		line := i + 2
		if len(rec) < 3 {
			return nil, fmt.Errorf("line %d: expected 3 columns, got %d", line, len(rec))
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid amount %q: %w", line, rec[1], err)
		}
		date, err := time.Parse(dateLayout, strings.TrimSpace(rec[2]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid date %q: %w", line, rec[2], err)
		}

		if !h.infra.TimeInRange(date, start, end) {
			continue
		}

		out = append(out, reconcile.BankStatement{
			UniqueID: strings.TrimSpace(rec[0]),
			Amount:   amount,
			Date:     date,
			Bank:     bankName,
		})
	}
	return out, nil
}

// writeReconcileSummary converts a usecase-layer summary into the
// HTTP-facing response type and writes it as JSON. Shared by both
// /reconcile and /reconcile/import so the response shape stays
// identical regardless of how the input was supplied.
func writeReconcileSummary(w http.ResponseWriter, summary reconcile.ReconcileSummary) {
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

	unmatchedAmount := make([]MatchedPair, 0, len(summary.AmountMismatch))

	for _, u := range summary.AmountMismatch {
		unmatchedAmount = append(unmatchedAmount, MatchedPair{
			System: SystemTransaction{
				TrxID:           u.System.TrxID,
				Type:            TrxType(u.System.Type),
				Amount:          u.System.Amount,
				TransactionTime: u.System.TransactionTime,
			},
			Bank: BankStatement{
				UniqueID: u.Bank.UniqueID,
				Amount:   u.Bank.Amount,
				Date:     u.Bank.Date,
				Bank:     u.Bank.Bank,
			},
			Discrepancy: u.Discrepancy,
		})
	}

	srvResult := ReconcileSummary{
		TotalProcessed:       summary.TotalProcessed,
		TotalMatched:         summary.TotalMatched,
		TotalUnmatched:       summary.TotalUnmatched,
		TotalUnmatchedAmount: summary.TotalUnmatchedAmount,
		TotalUnmatchedBank:   summary.TotalUnmatchedBank,
		TotalUnmatchedSystem: summary.TotalUnmatchedSystem,
		TotalDiscrepancy:     summary.TotalDiscrepancy,
		Matched:              matched,
		AmountMismatch:       unmatchedAmount,
		UnmatchedSystem:      unmatchedSystem,
		UnmatchedBank:        unmatchedBank,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(srvResult); err != nil {
		log.Printf("encoding response: %v", err)
	}
}

// writeError writes msg as a JSON error body with the given HTTP status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg})
}

// allowedBankNames returns allowedBanks' canonical display names,
// deduplicated and sorted, for use in error messages.
func allowedBankNames() string {
	seen := make(map[string]bool, len(allowedBanks))
	names := make([]string, 0, len(allowedBanks))
	for _, name := range allowedBanks {
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
