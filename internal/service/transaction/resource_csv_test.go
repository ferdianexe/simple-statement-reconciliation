package transaction

import (
	"context"
	"testing"
	"time"

	csv "github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestResource_ParseSystemTransactions(t *testing.T) {
	type mockFields struct {
		csvRepo *MockcsvRepoProvider
	}
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		mock    func(m mockFields)
		args    args
		want    []csv.Transaction
		wantErr error
	}{
		{
			name: "when_csvRepo_ParseSystemTransactions_return_non_nil_error_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.csvRepo.EXPECT().
					ParseSystemTransactions(context.Background(), "testdata/system_transactions.csv").
					Return(nil, assert.AnError)
			},
			args: args{
				ctx:  context.Background(),
				path: "testdata/system_transactions.csv",
			},
			want:    nil,
			wantErr: assert.AnError,
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.csvRepo.EXPECT().
					ParseSystemTransactions(context.Background(), "testdata/system_transactions.csv").
					Return([]csv.Transaction{
						{
							TrxID:           "TRX001",
							Amount:          110000,
							Type:            csv.Debit,
							TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
						},
					}, nil)
			},
			args: args{
				ctx:  context.Background(),
				path: "testdata/system_transactions.csv",
			},
			want: []csv.Transaction{
				{
					TrxID:           "TRX001",
					Amount:          110000,
					Type:            csv.Debit,
					TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
				},
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			csvRepoProvider := NewMockcsvRepoProvider(ctrl)

			mockFields := mockFields{
				csvRepo: csvRepoProvider,
			}
			test.mock(mockFields)

			rsc := &Resource{
				csvRepo: mockFields.csvRepo,
			}

			got, err := rsc.ParseSystemTransactions(test.args.ctx, test.args.path)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}
