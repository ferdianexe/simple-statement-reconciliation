package reconcile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/bank"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/service/transaction"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUsecase_Reconcile(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	type mockFields struct {
		bank        *MockbankServiceManager
		transaction *MocktransactionServiceManager
	}
	type args struct {
		ctx     context.Context
		request ReconcileRequest
	}
	tests := []struct {
		name    string
		mock    func(m mockFields)
		args    args
		want    ReconcileSummary
		wantErr error
	}{
		{
			name: "when_GetUserTransactionHistory_return_non_nil_error_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.transaction.EXPECT().
					GetUserTransactionHistory(context.Background(), transaction.TransactionHistoryParams{SysPath: "sys.csv"}).
					Return(nil, assert.AnError)
			},
			args: args{
				ctx: context.Background(),
				request: ReconcileRequest{
					SysPath: "sys.csv",
					Banks:   []BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
					Start:   start,
					End:     end,
				},
			},
			want:    ReconcileSummary{},
			wantErr: errors.New("parsing user transactions: assert.AnError general error for testing"),
		},
		{
			name: "when_GetBankStatementHistory_return_non_nil_error_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.transaction.EXPECT().
					GetUserTransactionHistory(context.Background(), transaction.TransactionHistoryParams{SysPath: "sys.csv"}).
					Return([]transaction.Transaction{}, nil)
				m.bank.EXPECT().
					GetBankStatementHistory(context.Background(), []bank.BankStatementParams{
						{BankName: "BCA", Path: "bca.csv", Start: start, End: end},
					}).
					Return(nil, assert.AnError)
			},
			args: args{
				ctx: context.Background(),
				request: ReconcileRequest{
					SysPath: "sys.csv",
					Banks:   []BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
					Start:   start,
					End:     end,
				},
			},
			want:    ReconcileSummary{},
			wantErr: errors.New("parsing bank statements: assert.AnError general error for testing"),
		},
		{
			name: "success_exact_match_no_discrepancy",
			mock: func(m mockFields) {
				m.transaction.EXPECT().
					GetUserTransactionHistory(context.Background(), transaction.TransactionHistoryParams{SysPath: "sys.csv"}).
					Return([]transaction.Transaction{
						{TrxID: "TRX1", Amount: 110000, Type: transaction.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
					}, nil)
				m.bank.EXPECT().
					GetBankStatementHistory(context.Background(), []bank.BankStatementParams{
						{BankName: "BCA", Path: "bca.csv", Start: start, End: end},
					}).
					Return([]bank.BankStatement{
						{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
					}, nil)
			},
			args: args{
				ctx: context.Background(),
				request: ReconcileRequest{
					SysPath: "sys.csv",
					Banks:   []BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
					Start:   start,
					End:     end,
				},
			},
			want: ReconcileSummary{
				TotalProcessed:   2,
				TotalMatched:     1,
				TotalUnmatched:   0,
				TotalDiscrepancy: 0,
				Matched: []MatchedPair{
					{
						System:      SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						Bank:        BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						Discrepancy: 0,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "success_matched_with_discrepancy",
			mock: func(m mockFields) {
				m.transaction.EXPECT().
					GetUserTransactionHistory(context.Background(), transaction.TransactionHistoryParams{SysPath: "sys.csv"}).
					Return([]transaction.Transaction{
						{TrxID: "TRX1", Amount: 110000, Type: transaction.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
					}, nil)
				m.bank.EXPECT().
					GetBankStatementHistory(context.Background(), []bank.BankStatementParams{
						{BankName: "BCA", Path: "bca.csv", Start: start, End: end},
					}).
					Return([]bank.BankStatement{
						{UniqueID: "BCA-1", Amount: -105000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
					}, nil)
			},
			args: args{
				ctx: context.Background(),
				request: ReconcileRequest{
					SysPath: "sys.csv",
					Banks:   []BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
					Start:   start,
					End:     end,
				},
			},
			want: ReconcileSummary{
				TotalProcessed:   2,
				TotalMatched:     1,
				TotalUnmatched:   0,
				TotalDiscrepancy: 5000,
				Matched: []MatchedPair{
					{
						System:      SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						Bank:        BankStatement{UniqueID: "BCA-1", Amount: -105000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						Discrepancy: 5000,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unmatched_system_transaction_when_no_bank_records_found",
			mock: func(m mockFields) {
				m.transaction.EXPECT().
					GetUserTransactionHistory(context.Background(), transaction.TransactionHistoryParams{SysPath: "sys.csv"}).
					Return([]transaction.Transaction{
						{TrxID: "TRX1", Amount: 110000, Type: transaction.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
					}, nil)
				m.bank.EXPECT().
					GetBankStatementHistory(context.Background(), []bank.BankStatementParams{
						{BankName: "BCA", Path: "bca.csv", Start: start, End: end},
					}).
					Return([]bank.BankStatement{}, nil)
			},
			args: args{
				ctx: context.Background(),
				request: ReconcileRequest{
					SysPath: "sys.csv",
					Banks:   []BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
					Start:   start,
					End:     end,
				},
			},
			want: ReconcileSummary{
				TotalProcessed:   1,
				TotalMatched:     0,
				TotalUnmatched:   1,
				TotalDiscrepancy: 0,
				UnmatchedSystem: []SystemTransaction{
					{TrxID: "TRX1", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
				},
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			bankService := NewMockbankServiceManager(ctrl)
			transactionService := NewMocktransactionServiceManager(ctrl)

			mockFields := mockFields{
				bank:        bankService,
				transaction: transactionService,
			}
			test.mock(mockFields)

			uc := &Usecase{
				bank:        mockFields.bank,
				transaction: mockFields.transaction,
			}

			got, err := uc.Reconcile(test.args.ctx, test.args.request)
			assert.Equal(t, test.want, got)
			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
