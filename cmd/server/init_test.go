package main

import (
	"testing"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPAppHandlers(t *testing.T) {
	type args struct {
		ucs *Usecases
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				ucs: NewUsecases(NewService(NewResources(csv.NewRepository()))),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewHTTPAppHandlers(test.args.ucs)

			assert.NotNil(t, got)
			assert.NotNil(t, got.HTTP.Reconcile)
		})
	}
}
