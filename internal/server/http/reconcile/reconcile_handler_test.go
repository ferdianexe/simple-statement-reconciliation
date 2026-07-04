package reconcile

import (
	"bytes"
	csvstd "encoding/csv"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gocsv"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/infrastructure/gotime"
	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func buildImportRequest(t *testing.T, start, end string, sysCSV *string, banks map[string]string) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	if start != "" {
		assert.NoError(t, mw.WriteField("start", start))
	}
	if end != "" {
		assert.NoError(t, mw.WriteField("end", end))
	}
	if sysCSV != nil {
		fw, err := mw.CreateFormFile(systemTransactionsField, "system_transactions.csv")
		assert.NoError(t, err)
		_, err = fw.Write([]byte(*sysCSV))
		assert.NoError(t, err)
	}
	for name, content := range banks {
		fw, err := mw.CreateFormFile(name, name+".csv")
		assert.NoError(t, err)
		_, err = fw.Write([]byte(content))
		assert.NoError(t, err)
	}
	assert.NoError(t, mw.Close())

	req := httptest.NewRequest(http.MethodPost, "/reconcile/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestHandler_HandleReconcileImport(t *testing.T) {
	type mockFields struct {
		reconcile *MockreconcileUsecaseManager
		infra     *MockinfraProvider
	}
	tests := []struct {
		name       string
		mock       func(m mockFields)
		buildReq   func(t *testing.T) *http.Request
		wantStatus int
		want       ReconcileSummary
		wantErrMsg string
	}{
		{
			name: "when_method_is_not_POST_then_return_405",
			mock: func(m mockFields) {},
			buildReq: func(t *testing.T) *http.Request {
				return httptest.NewRequest(http.MethodGet, "/reconcile/import", nil)
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantErrMsg: "only POST is supported",
		},
		{
			name: "when_multipart_form_is_invalid_then_return_400",
			mock: func(m mockFields) {},
			buildReq: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/reconcile/import", bytes.NewReader([]byte("not a multipart body")))
				req.Header.Set("Content-Type", "multipart/form-data; boundary=broken")
				return req
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid multipart form",
		},
		{
			name: "when_start_date_is_invalid_then_return_400",
			mock: func(m mockFields) {},
			buildReq: func(t *testing.T) *http.Request {
				return buildImportRequest(t, "not-a-date", "2024-01-31", nil, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid start date",
		},
		{
			name: "when_end_date_is_invalid_then_return_400",
			mock: func(m mockFields) {},
			buildReq: func(t *testing.T) *http.Request {
				return buildImportRequest(t, "2024-01-01", "not-a-date", nil, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid end date",
		},
		{
			name: "when_system_transactions_row_is_malformed_then_return_422",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
			},
			buildReq: func(t *testing.T) *http.Request {
				sysCSV := "trx_id,amount,type,transaction_time\nTRX1,not-a-number,DEBIT,2024-01-08T10:00:00Z\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", &sysCSV, nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErrMsg: "system_transactions",
		},
		{
			name: "when_system_transactions_row_has_invalid_type_then_return_422",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
			},
			buildReq: func(t *testing.T) *http.Request {
				sysCSV := "trx_id,amount,type,transaction_time\nTRX1,110000,UNKNOWN,2024-01-08T10:00:00Z\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", &sysCSV, nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErrMsg: "invalid type",
		},
		{
			name: "when_bank_file_row_is_malformed_then_return_422",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
			},
			buildReq: func(t *testing.T) *http.Request {
				bankCSV := "unique_identifier,amount,date\nBCA-1,-110000,not-a-date\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", nil, map[string]string{"BCA": bankCSV})
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErrMsg: "BCA",
		},
		{
			name: "when_bank_field_name_is_not_whitelisted_then_return_422",
			mock: func(m mockFields) {},
			buildReq: func(t *testing.T) *http.Request {
				bankCSV := "unique_identifier,amount,date\nXXX-1,-110000,2024-01-08\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", nil, map[string]string{"UNKNOWN_BANK": bankCSV})
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErrMsg: "unsupported bank",
		},
		{
			name: "when_bank_field_name_case_differs_from_whitelist_then_still_matches",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()

				m.reconcile.EXPECT().
					ReconcileFromRecords(
						gomock.Any(),
						[]reconcile.SystemTransaction(nil),
						[]reconcile.BankStatement{
							{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						},
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					).
					Return(reconcile.ReconcileSummary{}, nil)
			},
			buildReq: func(t *testing.T) *http.Request {
				bankCSV := "unique_identifier,amount,date\nBCA-1,-110000,2024-01-08\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", nil, map[string]string{"bca": bankCSV})
			},
			wantStatus: http.StatusOK,
			want: ReconcileSummary{
				Matched:         []MatchedPair{},
				UnmatchedSystem: []SystemTransaction{},
				UnmatchedBank:   []UnmatchedByBank{},
				AmountMismatch:  []MatchedPair{},
			},
		},
		{
			name: "when_system_transactions_file_is_missing_then_continue_with_empty_system_data",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()

				m.reconcile.EXPECT().
					ReconcileFromRecords(
						gomock.Any(),
						[]reconcile.SystemTransaction(nil),
						[]reconcile.BankStatement{
							{UniqueID: "BNI-1", Amount: 90000, Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Bank: "BNI"},
						},
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					).
					Return(reconcile.ReconcileSummary{}, nil)
			},
			buildReq: func(t *testing.T) *http.Request {
				bankCSV := "unique_identifier,amount,date\nBNI-1,90000,2024-01-03\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", nil, map[string]string{"BNI": bankCSV})
			},
			wantStatus: http.StatusOK,
			want: ReconcileSummary{
				Matched:         []MatchedPair{},
				UnmatchedSystem: []SystemTransaction{},
				UnmatchedBank:   []UnmatchedByBank{},
				AmountMismatch:  []MatchedPair{},
			},
		},
		{
			name: "when_no_bank_files_are_provided_then_continue_with_empty_bank_data",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()

				m.reconcile.EXPECT().
					ReconcileFromRecords(
						gomock.Any(),
						[]reconcile.SystemTransaction{
							{TrxID: "TRX1", Amount: 110000, Type: reconcile.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						},
						[]reconcile.BankStatement(nil),
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					).
					Return(reconcile.ReconcileSummary{}, nil)
			},
			buildReq: func(t *testing.T) *http.Request {
				sysCSV := "trx_id,amount,type,transaction_time\nTRX1,110000,DEBIT,2024-01-08T10:00:00Z\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", &sysCSV, nil)
			},
			wantStatus: http.StatusOK,
			want: ReconcileSummary{
				Matched:         []MatchedPair{},
				UnmatchedSystem: []SystemTransaction{},
				UnmatchedBank:   []UnmatchedByBank{},
				AmountMismatch:  []MatchedPair{},
			},
		},
		{
			name: "rows_outside_the_date_range_are_filtered_before_reaching_the_usecase",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()

				m.reconcile.EXPECT().
					ReconcileFromRecords(
						gomock.Any(),
						[]reconcile.SystemTransaction{
							{TrxID: "TRX1", Amount: 110000, Type: reconcile.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						},
						[]reconcile.BankStatement{
							{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						},
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					).
					Return(reconcile.ReconcileSummary{TotalProcessed: 2, TotalMatched: 1}, nil)
			},
			buildReq: func(t *testing.T) *http.Request {
				sysCSV := "trx_id,amount,type,transaction_time\n" +
					"TRX1,110000,DEBIT,2024-01-08T10:00:00Z\n" +
					"TRX2,50000,CREDIT,2023-12-25T09:00:00Z\n" // outside range
				bankCSV := "unique_identifier,amount,date\n" +
					"BCA-1,-110000,2024-01-08\n" +
					"BCA-2,-99999,2024-02-05\n" // outside range
				return buildImportRequest(t, "2024-01-01", "2024-01-31", &sysCSV, map[string]string{"BCA": bankCSV})
			},
			wantStatus: http.StatusOK,
			want: ReconcileSummary{
				TotalProcessed:  2,
				TotalMatched:    1,
				Matched:         []MatchedPair{},
				UnmatchedSystem: []SystemTransaction{},
				UnmatchedBank:   []UnmatchedByBank{},
				AmountMismatch:  []MatchedPair{},
			},
		},
		{
			name: "when_usecase_ReconcileFromRecords_return_non_nil_error_then_return_422",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()

				m.reconcile.EXPECT().
					ReconcileFromRecords(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					).
					Return(reconcile.ReconcileSummary{}, assert.AnError)
			},
			buildReq: func(t *testing.T) *http.Request {
				sysCSV := "trx_id,amount,type,transaction_time\nTRX1,110000,DEBIT,2024-01-08T10:00:00Z\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", &sysCSV, nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErrMsg: assert.AnError.Error(),
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.infra.EXPECT().CsvNewReader(gomock.Any()).DoAndReturn(func(r io.Reader) *csvstd.Reader {
					return gocsv.Default.NewReader(r)
				}).AnyTimes()
				m.infra.EXPECT().CsvReadAll(gomock.Any()).DoAndReturn(func(r gocsv.Reader) ([][]string, error) {
					return gocsv.Default.ReadAll(r)
				}).AnyTimes()
				m.infra.EXPECT().TimeInRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(d, start, end time.Time) bool {
					return gotime.Default.InRange(d, start, end)
				}).AnyTimes()
				m.reconcile.EXPECT().
					ReconcileFromRecords(
						gomock.Any(),
						[]reconcile.SystemTransaction{
							{TrxID: "TRX1", Amount: 110000, Type: reconcile.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						},
						[]reconcile.BankStatement{
							{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						},
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					).
					Return(reconcile.ReconcileSummary{
						TotalProcessed:       2,
						TotalMatched:         1,
						TotalUnmatched:       0,
						TotalDiscrepancy:     0,
						TotalUnmatchedAmount: 1,
						TotalUnmatchedBank:   1,
						TotalUnmatchedSystem: 1,
						Matched: []reconcile.MatchedPair{
							{
								System:      reconcile.SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: reconcile.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
								Bank:        reconcile.BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
								Discrepancy: 0,
							},
						},
						AmountMismatch: []reconcile.MatchedPair{
							{
								System:      reconcile.SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: reconcile.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
								Bank:        reconcile.BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
								Discrepancy: 1,
							},
						},
						UnmatchedSystem: []reconcile.SystemTransaction{
							{
								TrxID:           "TRX69",
								Amount:          10000,
								Type:            reconcile.Debit,
								TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
							},
						},
						UnmatchedBank: []reconcile.UnmatchedByBank{
							{
								Bank: "BCA",
								Records: []reconcile.BankStatement{
									{
										UniqueID: "BCA-6",
										Amount:   69,
										Date:     time.Time{},
										Bank:     "BCA",
									},
								},
							},
						},
					}, nil)
			},
			buildReq: func(t *testing.T) *http.Request {
				sysCSV := "trx_id,amount,type,transaction_time\nTRX1,110000,DEBIT,2024-01-08T10:00:00Z\n"
				bankCSV := "unique_identifier,amount,date\nBCA-1,-110000,2024-01-08\n"
				return buildImportRequest(t, "2024-01-01", "2024-01-31", &sysCSV, map[string]string{"BCA": bankCSV})
			},
			wantStatus: http.StatusOK,
			want: ReconcileSummary{
				TotalProcessed:       2,
				TotalMatched:         1,
				TotalUnmatched:       0,
				TotalDiscrepancy:     0,
				TotalUnmatchedAmount: 1,
				TotalUnmatchedBank:   1,
				TotalUnmatchedSystem: 1,
				Matched: []MatchedPair{
					{
						System:      SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						Bank:        BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						Discrepancy: 0,
					},
				},
				AmountMismatch: []MatchedPair{
					{
						System:      SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						Bank:        BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						Discrepancy: 1,
					},
				},
				UnmatchedSystem: []SystemTransaction{
					{
						TrxID:           "TRX69",
						Amount:          10000,
						Type:            Debit,
						TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
					},
				},
				UnmatchedBank: []UnmatchedByBank{
					{
						Bank: "BCA",
						Records: []BankStatement{
							{
								UniqueID: "BCA-6",
								Amount:   69,
								Date:     time.Time{},
								Bank:     "BCA",
							},
						},
					},
				},
			},
			wantErrMsg: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			reconcileUsecase := NewMockreconcileUsecaseManager(ctrl)
			infra := NewMockinfraProvider(ctrl)

			mockFields := mockFields{
				reconcile: reconcileUsecase,
				infra:     infra,
			}
			test.mock(mockFields)

			h := &Handler{
				reconcile: mockFields.reconcile,
				infra:     mockFields.infra,
			}

			req := test.buildReq(t)
			rec := httptest.NewRecorder()

			h.HandleReconcileImport(rec, req)

			assert.Equal(t, test.wantStatus, rec.Code)

			if test.wantStatus == http.StatusOK {
				var got ReconcileSummary
				err := json.Unmarshal(rec.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, test.want, got)
				return
			}

			var errBody errorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errBody)
			assert.NoError(t, err)
			assert.Contains(t, errBody.Error, test.wantErrMsg)
		})
	}
}
