package main

import "github.com/ferdianexe/simple-statement-reconciliation/internal/server/http/reconcile"

// HTTPAppHandlers contains list of http and nsq handler to support main app.
type HTTPAppHandlers struct {
	HTTP HTTPHandlers
}

// HTTPHandlers contains a list of http handler to support main app.
type HTTPHandlers struct {
	Reconcile *reconcile.Handler
}

// NewHTTPAppHandlers initialize handlers with each use case according to the requirement for each handlers.
// It accepts ucs, writer, infra, featureFlag as parameters.
// It returns non nil pointer HTTPAppHandlers.
func NewHTTPAppHandlers(ucs *Usecases) *HTTPAppHandlers {
	h := &HTTPAppHandlers{
		HTTP: HTTPHandlers{
			Reconcile: reconcile.NewHandler(ucs.reconcile),
		},
	}

	return h
}
