package main

import (
	"testing"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPAppHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ucs   *Usecases
		infra infraProvider
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				ucs:   NewUsecases(NewService(NewResources(csv.NewRepository(NewMockinfraProvider(ctrl)))), NewMockinfraProvider(ctrl)),
				infra: NewMockinfraProvider(ctrl),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewHTTPAppHandlers(test.args.ucs, test.args.infra)

			assert.NotNil(t, got)
			assert.NotNil(t, got.HTTP.Reconcile)
		})
	}
}
