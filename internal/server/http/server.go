package http

import (
	"log"
	"net/http"
)

//go:generate mockgen -source=server.go -destination=server_mock.go -package=http

type reconcileHandler interface {
	HandleReconcile(w http.ResponseWriter, r *http.Request)
	// HandleReconcileImport handles the CSV-upload variant of reconciliation.
	HandleReconcileImport(w http.ResponseWriter, r *http.Request)
}

// Handlers holds a collection of HTTP handlers provided for server.
type Handlers struct {
	Reconcile reconcileHandler
}

// Server type of HTTP server.
type Server struct {
	handlers Handlers
}

// NewServer initialize server with config and handler as parameters.
func NewServer(handlers Handlers) *Server {
	return &Server{
		handlers: handlers,
	}
}

// NewMux builds the HTTP routes for the service. Exposed separately from
// Run so tests can exercise handlers with httptest.NewServer without
// binding a real port.
func (srv *Server) NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/reconcile", srv.handlers.Reconcile.HandleReconcile)
	mux.HandleFunc("/reconcile/import", srv.handlers.Reconcile.HandleReconcileImport)
	return mux
}

// Run starts the HTTP server and blocks until it exits.
func (srv *Server) Run(addr string) error {
	log.Printf("reconciliation server listening on %s", addr)
	return http.ListenAndServe(addr, srv.NewMux())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
