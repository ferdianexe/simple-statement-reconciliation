package bank

import (
	"context"
	"errors"
	"testing"
	"time"

	csv "github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_GetBankStatementHistory(t *testing.T) {
	type mockFields struct {
		resource *MockresourceProvider
	}
	type args struct {
		ctx     context.Context
		request []BankStatementParams
	}
	tests := []struct {
		name    string
		mock    func(m mockFields)
		args    args
		want    []BankStatement
		wantErr error
	}{
		{
			name: "when_ParseBankStatement_return_non_nil_error_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.resource.EXPECT().
					ParseBankStatementFromCSV(context.Background(), "testdata/bank.csv", "BCA", time.Time{}, time.Time{}).
					Return(nil, assert.AnError)
			},
			args: args{
				ctx: context.Background(),
				request: []BankStatementParams{
					{BankName: "BCA", Path: "testdata/bank.csv"},
				},
			},
			want:    []BankStatement{},
			wantErr: errors.New("parsing bank statement testdata/bank.csv: assert.AnError general error for testing"),
		},
		{
			name: "when_path_is_empty_then_derive_default_path_from_bank_name",
			mock: func(m mockFields) {
				m.resource.EXPECT().
					ParseBankStatementFromCSV(context.Background(), "testdata/bank_bca.csv", "BCA", time.Time{}, time.Time{}).
					Return([]csv.BankStatement{
						{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
					}, nil)
			},
			args: args{
				ctx: context.Background(),
				request: []BankStatementParams{
					{BankName: "BCA", Path: ""},
				},
			},
			want: []BankStatement{
				{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
			},
			wantErr: nil,
		},
		{
			name: "success_with_multiple_banks_aggregates_results",
			mock: func(m mockFields) {
				m.resource.EXPECT().
					ParseBankStatementFromCSV(context.Background(), "testdata/bca.csv", "BCA", time.Time{}, time.Time{}).
					Return([]csv.BankStatement{
						{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
					}, nil)
				m.resource.EXPECT().
					ParseBankStatementFromCSV(context.Background(), "testdata/bni.csv", "BNI", time.Time{}, time.Time{}).
					Return([]csv.BankStatement{
						{UniqueID: "BNI-1", Amount: 220000, Date: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Bank: "BNI"},
					}, nil)
			},
			args: args{
				ctx: context.Background(),
				request: []BankStatementParams{
					{BankName: "BCA", Path: "testdata/bca.csv"},
					{BankName: "BNI", Path: "testdata/bni.csv"},
				},
			},
			want: []BankStatement{
				{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
				{UniqueID: "BNI-1", Amount: 220000, Date: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Bank: "BNI"},
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

			got, err := svc.GetBankStatementHistory(test.args.ctx, test.args.request)
			assert.Equal(t, test.want, got)
			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
