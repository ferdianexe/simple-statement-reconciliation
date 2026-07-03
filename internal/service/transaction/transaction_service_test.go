package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	csv "github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_GetUserTransactionHistory(t *testing.T) {
	type mockFields struct {
		resource *MockresourceProvider
	}
	type args struct {
		ctx     context.Context
		request TransactionHistoryParams
	}
	tests := []struct {
		name    string
		mock    func(m mockFields)
		args    args
		want    []Transaction
		wantErr error
	}{
		{
			name: "when_ParseSystemTransactions_return_non_nil_error_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.resource.EXPECT().
					ParseSystemTransactions(context.Background(), "testdata/system_transactions.csv").
					Return(nil, assert.AnError)
			},
			args: args{
				ctx:     context.Background(),
				request: TransactionHistoryParams{SysPath: "testdata/system_transactions.csv"},
			},
			want:    []Transaction{},
			wantErr: errors.New("parsing system transactions: assert.AnError general error for testing"),
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.resource.EXPECT().
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
				ctx:     context.Background(),
				request: TransactionHistoryParams{SysPath: "testdata/system_transactions.csv"},
			},
			want: []Transaction{
				{
					TrxID:           "TRX001",
					Amount:          110000,
					Type:            Debit,
					TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
				},
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resourceProvider := NewMockresourceProvider(ctrl)

			mockFields := mockFields{
				resource: resourceProvider,
			}
			test.mock(mockFields)

			svc := &Service{
				resource: mockFields.resource,
			}

			got, err := svc.GetUserTransactionHistory(test.args.ctx, test.args.request)
			assert.Equal(t, test.want, got)
			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
