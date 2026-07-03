package bank

import (
	"context"
	"testing"
	"time"

	csv "github.com/ferdianexe/simple-statement-reconciliation/internal/repository/csv"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestResource_ParseBankStatement(t *testing.T) {
	type mockFields struct {
		csvRepo *MockcsvRepoProvider
	}
	type args struct {
		ctx      context.Context
		path     string
		bankName string
	}
	tests := []struct {
		name    string
		mock    func(m mockFields)
		args    args
		want    []csv.BankStatement
		wantErr error
	}{
		{
			name: "when_csvRepo_ParseBankStatement_return_non_nil_error_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.csvRepo.EXPECT().
					ParseBankStatement(context.Background(), "testdata/bank_bca.csv", "BCA").
					Return(nil, assert.AnError)
			},
			args: args{
				ctx:      context.Background(),
				path:     "testdata/bank_bca.csv",
				bankName: "BCA",
			},
			want:    nil,
			wantErr: assert.AnError,
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.csvRepo.EXPECT().
					ParseBankStatement(context.Background(), "testdata/bank_bca.csv", "BCA").
					Return([]csv.BankStatement{
						{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
					}, nil)
			},
			args: args{
				ctx:      context.Background(),
				path:     "testdata/bank_bca.csv",
				bankName: "BCA",
			},
			want: []csv.BankStatement{
				{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
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

			got, err := rsc.ParseBankStatement(test.args.ctx, test.args.path, test.args.bankName)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}
