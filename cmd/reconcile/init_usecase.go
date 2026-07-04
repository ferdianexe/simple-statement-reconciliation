package main

import "github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"

// Usecases is a collection of usecase to initialize the HTTP app.
type Usecases struct {
	reconcile *reconcile.Usecase
}

// NewUsecases initialize usecases with each service according to the requirement for each usecases.
func NewUsecases(services *Services, infra infraProvider) *Usecases {
	usecases := &Usecases{
		reconcile: reconcile.NewUsecase(services.bank, services.transaction, infra),
	}
	return usecases
}
