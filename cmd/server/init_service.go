package main

import (
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/bank"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/transaction"
)

// Services is a collection of usecase to initialize the HTTP app.
type Services struct {
	bank        *bank.Service
	transaction *transaction.Service
}

// NewService initialize service with each service according to the requirement for each usecases.
func NewService(resource *Resources) *Services {
	usecases := &Services{
		bank:        bank.NewService(resource.bank),
		transaction: transaction.NewService(resource.transaction),
	}
	return usecases
}
