// Command server runs the reconciliation service as an HTTP server.
//
// Example:
//
//	go run ./cmd/server -addr :8080
//
//	curl -X POST localhost:8080/reconcile -d '{
//	  "sys_path": "testdata/system_transactions.csv",
//	  "banks": [
//	    {"name": "BCA", "path": "testdata/bank_bca.csv"},
//	    {"name": "BNI", "path": "testdata/bank_bni.csv"}
//	  ],
//	  "start": "2024-01-01",
//	  "end": "2024-01-31"
//	}'
package main

import (
	"flag"
	"log"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gotime"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/server/http"
)

func main() {
	addr := flag.String("addr", ":8080", "address for the HTTP server to listen on")
	flag.Parse()
	infra := infrastructure.NewService(infrastructure.NewServiceParam{
		Csv:  gocsv.Default,
		Time: gotime.Default,
	})
	csvRepo := csv.NewRepository(infra)
	rsc := NewResources(csvRepo)
	services := NewService(rsc)
	usecaseApp := NewUsecases(services, infra)
	httpAppHandlers := NewHTTPAppHandlers(usecaseApp, infra)
	srv := http.NewServer(http.Handlers{
		Reconcile: httpAppHandlers.HTTP.Reconcile,
	})
	if err := srv.Run(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
