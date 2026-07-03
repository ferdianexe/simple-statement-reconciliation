package csv

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// writeTempCSV writes content to a temp file and returns its path. Used
// instead of gomock in this package: Repository has no injected
// dependencies to mock (it's the layer that actually does file I/O), so
// the loop still follows the project's table-driven convention, just
// against real files instead of mocked collaborators.
func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp csv: %v", err)
	}
	return path
}

func TestRepository_ParseSystemTransactions(t *testing.T) {
	type args struct {
		ctx     context.Context
		content string
	}
	tests := []struct {
		name      string
		args      args
		want      []Transaction
		wantErr   bool
		wantErrIs string
	}{
		{
			name: "when_file_does_not_exist_then_return_non_nil_error",
			args: args{
				ctx:     context.Background(),
				content: "__missing__", // handled specially below
			},
			wantErr:   true,
			wantErrIs: "open system transactions file",
		},
		{
			name: "when_amount_is_invalid_then_return_non_nil_error",
			args: args{
				ctx:     context.Background(),
				content: "trx_id,amount,type,transaction_time\nTRX001,not-a-number,DEBIT,2024-01-08T10:00:00Z\n",
			},
			wantErr:   true,
			wantErrIs: "invalid amount",
		},
		{
			name: "when_type_is_invalid_then_return_non_nil_error",
			args: args{
				ctx:     context.Background(),
				content: "trx_id,amount,type,transaction_time\nTRX001,110000,UNKNOWN,2024-01-08T10:00:00Z\n",
			},
			wantErr:   true,
			wantErrIs: "invalid type",
		},
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				content: "trx_id,amount,type,transaction_time\nTRX001,110000,DEBIT,2024-01-08T10:00:00Z\n",
			},
			want: []Transaction{
				{
					TrxID:           "TRX001",
					Amount:          110000,
					Type:            Debit,
					TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := NewRepository()

			path := test.args.content
			if path != "__missing__" {
				path = writeTempCSV(t, test.args.content)
			} else {
				path = filepath.Join(t.TempDir(), "does-not-exist.csv")
			}

			got, err := repo.ParseSystemTransactions(test.args.ctx, path)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErrIs)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestRepository_ParseBankStatement(t *testing.T) {
	type args struct {
		ctx      context.Context
		content  string
		bankName string
	}
	tests := []struct {
		name      string
		args      args
		want      []BankStatement
		wantErr   bool
		wantErrIs string
	}{
		{
			name: "when_file_does_not_exist_then_return_non_nil_error",
			args: args{
				ctx:      context.Background(),
				content:  "__missing__",
				bankName: "BCA",
			},
			wantErr:   true,
			wantErrIs: "open bank statement file",
		},
		{
			name: "when_amount_is_invalid_then_return_non_nil_error",
			args: args{
				ctx:      context.Background(),
				content:  "unique_identifier,amount,date\nBCA-1,not-a-number,2024-01-08\n",
				bankName: "BCA",
			},
			wantErr:   true,
			wantErrIs: "invalid amount",
		},
		{
			name: "when_date_is_invalid_then_return_non_nil_error",
			args: args{
				ctx:      context.Background(),
				content:  "unique_identifier,amount,date\nBCA-1,-110000,not-a-date\n",
				bankName: "BCA",
			},
			wantErr:   true,
			wantErrIs: "invalid date",
		},
		{
			name: "success",
			args: args{
				ctx:      context.Background(),
				content:  "unique_identifier,amount,date\nBCA-1,-110000,2024-01-08\n",
				bankName: "BCA",
			},
			want: []BankStatement{
				{
					UniqueID: "BCA-1",
					Amount:   -110000,
					Date:     time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
					Bank:     "BCA",
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := NewRepository()

			path := test.args.content
			if path != "__missing__" {
				path = writeTempCSV(t, test.args.content)
			} else {
				path = filepath.Join(t.TempDir(), "does-not-exist.csv")
			}

			got, err := repo.ParseBankStatement(test.args.ctx, path, test.args.bankName)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErrIs)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
