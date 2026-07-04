package csv

import (
	"context"
	csvstd "encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gotime"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// writeTempCSV writes content to a temp file and returns its path. The
// repository still opens the file itself (os.Open isn't part of the
// injected infra), so a real file must exist even though CsvReadAll is
// mocked and the file's actual content is irrelevant in most cases.
func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp csv: %v", err)
	}
	return path
}

// dummyReader stands in for whatever CsvNewReader would have returned;
// since CsvReadAll is mocked separately to return canned records, its
// actual content is never touched.
func dummyReader() *csvstd.Reader {
	return csvstd.NewReader(strings.NewReader(""))
}

func TestRepository_ParseSystemTransactions(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	type mockFields struct {
		infra *MockinfraProvider
	}
	type args struct {
		ctx        context.Context
		path       string
		start, end time.Time
	}
	tests := []struct {
		name      string
		mock      func(m mockFields)
		args      func(t *testing.T) args
		want      []Transaction
		wantErr   bool
		wantErrIs string
	}{
		{
			name: "when_file_does_not_exist_then_return_non_nil_error",
			mock: func(m mockFields) {},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: filepath.Join(t.TempDir(), "does-not-exist.csv"), start: start, end: end}
			},
			wantErr:   true,
			wantErrIs: "open system transactions file",
		},
		{
			name: "when_CsvReadAll_returns_non_nil_error_then_return_it_wrapped",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return(nil, assert.AnError)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), start: start, end: end}
			},
			wantErr:   true,
			wantErrIs: assert.AnError.Error(),
		},
		{
			name: "when_records_are_empty_then_return_nil_no_error",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return(nil, nil)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), start: start, end: end}
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "when_type_is_invalid_then_return_non_nil_error",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return([][]string{
					{"trx_id", "amount", "type", "transaction_time"},
					{"TRX001", "110000", "UNKNOWN", "2024-01-08T10:00:00Z"},
				}, nil)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), start: start, end: end}
			},
			wantErr:   true,
			wantErrIs: "line 2: invalid type",
		},
		{
			name: "when_row_is_outside_date_range_then_it_is_filtered_out",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return([][]string{
					{"trx_id", "amount", "type", "transaction_time"},
					{"TRX001", "110000", "DEBIT", "2024-01-08T10:00:00Z"}, // in range
					{"TRX002", "50000", "CREDIT", "2023-12-25T09:00:00Z"}, // before range
				}, nil)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), start: start, end: end}
			},
			want: []Transaction{
				{TrxID: "TRX001", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
			},
			wantErr: false,
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return([][]string{
					{"trx_id", "amount", "type", "transaction_time"},
					{"TRX001", "110000", "DEBIT", "2024-01-08T10:00:00Z"},
				}, nil)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), start: start, end: end}
			},
			want: []Transaction{
				{TrxID: "TRX001", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockFields := mockFields{infra: NewMockinfraProvider(ctrl)}
			test.mock(mockFields)

			repo := &Repository{infra: mockFields.infra}

			a := test.args(t)
			got, err := repo.ParseSystemTransactions(a.ctx, a.path, a.start, a.end)

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
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	type mockFields struct {
		infra *MockinfraProvider
	}
	type args struct {
		ctx        context.Context
		path       string
		bankName   string
		start, end time.Time
	}
	tests := []struct {
		name      string
		mock      func(m mockFields)
		args      func(t *testing.T) args
		want      []BankStatement
		wantErr   bool
		wantErrIs string
	}{
		{
			name: "when_file_does_not_exist_then_return_non_nil_error",
			mock: func(m mockFields) {},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: filepath.Join(t.TempDir(), "does-not-exist.csv"), bankName: "BCA", start: start, end: end}
			},
			wantErr:   true,
			wantErrIs: "open bank statement file",
		},
		{
			name: "when_CsvReadAll_returns_non_nil_error_then_wrap_with_bank_name",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return(nil, assert.AnError)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), bankName: "BCA", start: start, end: end}
			},
			wantErr:   true,
			wantErrIs: "BCA: read csv: " + assert.AnError.Error(),
		},
		{
			name: "when_row_is_outside_date_range_then_it_is_filtered_out",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return([][]string{
					{"unique_identifier", "amount", "date"},
					{"BCA-1", "-110000", "2024-01-08"}, // in range
					{"BCA-2", "-50000", "2024-02-05"},  // after range
				}, nil)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), bankName: "BCA", start: start, end: end}
			},
			want: []BankStatement{
				{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
			},
			wantErr: false,
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.infra.EXPECT().TimeTruncateToDay(gomock.Any()).DoAndReturn(func(t time.Time) time.Time {
					return gotime.Default.TruncateToDay(t)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.infra.EXPECT().CsvNewReader(gomock.Any()).Return(dummyReader())
				m.infra.EXPECT().CsvReadAll(gomock.Any()).Return([][]string{
					{"unique_identifier", "amount", "date"},
					{"BCA-1", "-110000", "2024-01-08"},
				}, nil)
			},
			args: func(t *testing.T) args {
				return args{ctx: context.Background(), path: writeTempCSV(t, "irrelevant, reading is mocked"), bankName: "BCA", start: start, end: end}
			},
			want: []BankStatement{
				{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockFields := mockFields{infra: NewMockinfraProvider(ctrl)}
			test.mock(mockFields)

			repo := &Repository{infra: mockFields.infra}

			a := test.args(t)
			got, err := repo.ParseBankStatement(a.ctx, a.path, a.bankName, a.start, a.end)

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
