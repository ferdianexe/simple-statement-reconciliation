package main

import (
	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/bank"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/transaction"
)

// Resources is a collection of usecase to initialize the HTTP app.
type Resources struct {
	bank        *bank.Resource
	transaction *transaction.Resource
}

// NewUsecases initialize usecases with each service according to the requirement for each usecases.
func NewResources(csvRepo *csv.Repository) *Resources {
	resources := &Resources{
		bank:        bank.NewResource(csvRepo),
		transaction: transaction.NewResource(csvRepo),
	}
	return resources
}
