package main

import (
	"testing"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	"github.com/stretchr/testify/assert"
)

func TestNewResources(t *testing.T) {
	type args struct {
		csvRepo *csv.Repository
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				csvRepo: csv.NewRepository(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewResources(test.args.csvRepo)

			assert.NotNil(t, got)
			assert.NotNil(t, got.bank)
			assert.NotNil(t, got.transaction)
		})
	}
}
