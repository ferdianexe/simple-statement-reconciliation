package main

import (
	"testing"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/stretchr/testify/assert"
)

func TestNewUsecases(t *testing.T) {
	type args struct {
		services *Services
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				services: NewService(NewResources(csv.NewRepository())),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewUsecases(test.args.services)

			assert.NotNil(t, got)
			assert.NotNil(t, got.reconcile)
		})
	}
}
