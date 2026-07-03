package main

import (
	"testing"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	type args struct {
		resource *Resources
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				resource: NewResources(csv.NewRepository()),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewService(test.args.resource)

			assert.NotNil(t, got)
			assert.NotNil(t, got.bank)
			assert.NotNil(t, got.transaction)
		})
	}
}
