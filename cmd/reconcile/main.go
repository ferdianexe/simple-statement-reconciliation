// Command reconcile runs the transaction reconciliation service as a
// one-off batch CLI (as opposed to cmd/server, which runs the same
// underlying logic behind an HTTP API).
//
// Example:
//
//	go run ./cmd/reconcile \
//	  -sys testdata/system_transactions.csv \
//	  -banks "BCA:testdata/bank_bca.csv,BNI:testdata/bank_bni.csv" \
//	  -start 2024-01-01 -end 2024-01-31
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gotime"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
)

const dateLayout = "2006-01-02"

func main() {
	sysPath := flag.String("sys", "", "path to system transactions CSV (required)")
	banksFlag := flag.String("banks", "", `comma-separated "BankName:path.csv" pairs (required), e.g. "BCA:bca.csv,BNI:bni.csv"`)
	startStr := flag.String("start", "", "reconciliation start date, YYYY-MM-DD (required)")
	endStr := flag.String("end", "", "reconciliation end date, YYYY-MM-DD (required)")
	asJSON := flag.Bool("json", false, "print the summary as JSON instead of plain text")
	flag.Parse()

	if *sysPath == "" || *banksFlag == "" || *startStr == "" || *endStr == "" {
		flag.Usage()
		os.Exit(2)
	}

	banks, err := parseBanksFlag(*banksFlag)
	if err != nil {
		log.Fatalf("invalid -banks: %v", err)
	}

	start, err := time.Parse(dateLayout, *startStr)
	if err != nil {
		log.Fatalf("invalid -start date: %v", err)
	}
	end, err := time.Parse(dateLayout, *endStr)
	if err != nil {
		log.Fatalf("invalid -end date: %v", err)
	}

	infra := infrastructure.NewService(infrastructure.NewServiceParam{
		Csv:  gocsv.Default,
		Time: gotime.Default,
	})
	csvRepo := csv.NewRepository(infra)
	rsc := NewResources(csvRepo)
	services := NewService(rsc)
	usecaseApp := NewUsecases(services, infra)

	summary, err := usecaseApp.reconcile.Reconcile(context.Background(), reconcile.ReconcileRequest{
		SysPath: *sysPath,
		Banks:   banks,
		Start:   start,
		End:     end,
	})
	if err != nil {
		log.Fatalf("reconcile: %v", err)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(summary); err != nil {
			log.Fatalf("encoding summary: %v", err)
		}
		return
	}

	printSummary(summary)
}

// parseBanksFlag turns "BCA:path.csv,BNI:path2.csv" into []service.BankFile.
// This string format is a CLI-only convenience; the HTTP API (cmd/server)
// takes the equivalent structured JSON directly.
func parseBanksFlag(raw string) ([]reconcile.BankFileRequest, error) {
	var banks []reconcile.BankFileRequest
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid entry %q, expected BankName:path.csv", pair)
		}
		banks = append(banks, reconcile.BankFileRequest{Name: parts[0], Path: parts[1]})
	}
	return banks, nil
}

func printSummary(s reconcile.ReconcileSummary) {
	fmt.Printf("Total processed:   %d\n", s.TotalProcessed)
	fmt.Printf("Total matched:     %d\n", s.TotalMatched)
	fmt.Printf("Total unmatched:   %d\n", s.TotalUnmatched)
	fmt.Printf("Total discrepancy: %.2f\n\n", s.TotalDiscrepancy)

	if len(s.UnmatchedSystem) > 0 {
		fmt.Println("Unmatched system transactions:")
		for _, t := range s.UnmatchedSystem {
			fmt.Printf("  - %s | %.2f | %s | %s\n", t.TrxID, t.Amount, t.Type, t.TransactionTime.Format(time.RFC3339))
		}
		fmt.Println()
	}

	if len(s.UnmatchedBank) > 0 {
		fmt.Println("Unmatched bank statement records:")
		for _, group := range s.UnmatchedBank {
			fmt.Printf("  %s:\n", group.Bank)
			for _, r := range group.Records {
				fmt.Printf("    - %s | %.2f | %s\n", r.UniqueID, r.Amount, r.Date.Format(dateLayout))
			}
		}
	}
}
