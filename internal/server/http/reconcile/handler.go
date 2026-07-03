package reconcile

import (
	"context"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
)

//go:generate mockgen -source=handler.go -destination=mock_handler.go -package=reconcile

// reconcileUsecaseManager provides the required action usecase methods.
type reconcileUsecaseManager interface {
	// Reconcile performs reconciliation process.
	Reconcile(ctx context.Context, request reconcile.ReconcileRequest) (reconcile.ReconcileSummary, error)
}

// Handler type of reconcile handler.
type Handler struct {
	reconcile reconcileUsecaseManager
}

// NewHandler instantiates reconcile handler.
func NewHandler(reconcile reconcileUsecaseManager) *Handler {
	return &Handler{
		reconcile: reconcile,
	}
}
