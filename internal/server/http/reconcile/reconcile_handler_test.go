package reconcile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"encoding/json"

	"github.com/ferdianexe/simple-statement-reconciliation/internal/usecase/reconcile"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_HandleReconcile(t *testing.T) {
	type mockFields struct {
		reconcile *MockreconcileUsecaseManager
	}
	type args struct {
		method string
		body   string
	}
	tests := []struct {
		name       string
		mock       func(m mockFields)
		args       args
		wantStatus int
		want       ReconcileSummary // only asserted when wantStatus == http.StatusOK
		wantErrMsg string           // substring expected in the error body otherwise
	}{
		{
			name: "when_method_is_not_POST_then_return_405",
			mock: func(m mockFields) {},
			args: args{
				method: http.MethodGet,
				body:   "",
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantErrMsg: "only POST is supported",
		},
		{
			name: "when_body_is_invalid_json_then_return_400",
			mock: func(m mockFields) {},
			args: args{
				method: http.MethodPost,
				body:   "{not-json",
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid JSON body",
		},
		{
			name: "when_start_date_is_invalid_then_return_400",
			mock: func(m mockFields) {},
			args: args{
				method: http.MethodPost,
				body:   `{"sys_path":"sys.csv","banks":[{"name":"BCA","path":"bca.csv"}],"start":"not-a-date","end":"2024-01-31"}`,
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid start date",
		},
		{
			name: "when_end_date_is_invalid_then_return_400",
			mock: func(m mockFields) {},
			args: args{
				method: http.MethodPost,
				body:   `{"sys_path":"sys.csv","banks":[{"name":"BCA","path":"bca.csv"}],"start":"2024-01-01","end":"not-a-date"}`,
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid end date",
		},
		{
			name: "when_usecase_Reconcile_return_non_nil_error_then_return_422",
			mock: func(m mockFields) {
				m.reconcile.EXPECT().
					Reconcile(context.Background(), reconcile.ReconcileRequest{
						SysPath: "sys.csv",
						Banks:   []reconcile.BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
						Start:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						End:     time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					}).
					Return(reconcile.ReconcileSummary{}, assert.AnError)
			},
			args: args{
				method: http.MethodPost,
				body:   `{"sys_path":"sys.csv","banks":[{"name":"BCA","path":"bca.csv"}],"start":"2024-01-01","end":"2024-01-31"}`,
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErrMsg: assert.AnError.Error(),
		},
		{
			name: "success",
			mock: func(m mockFields) {
				m.reconcile.EXPECT().
					Reconcile(context.Background(), reconcile.ReconcileRequest{
						SysPath: "sys.csv",
						Banks:   []reconcile.BankFileRequest{{Name: "BCA", Path: "bca.csv"}},
						Start:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						End:     time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					}).
					Return(reconcile.ReconcileSummary{
						TotalProcessed:   2,
						TotalMatched:     1,
						TotalUnmatched:   0,
						TotalDiscrepancy: 0,
						Matched: []reconcile.MatchedPair{
							{
								System:      reconcile.SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: reconcile.Debit, TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
								Bank:        reconcile.BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
								Discrepancy: 0,
							},
						},
					}, nil)
			},
			args: args{
				method: http.MethodPost,
				body:   `{"sys_path":"sys.csv","banks":[{"name":"BCA","path":"bca.csv"}],"start":"2024-01-01","end":"2024-01-31"}`,
			},
			wantStatus: http.StatusOK,
			want: ReconcileSummary{
				TotalProcessed:   2,
				TotalMatched:     1,
				TotalUnmatched:   0,
				TotalDiscrepancy: 0,
				Matched: []MatchedPair{
					{
						System:      SystemTransaction{TrxID: "TRX1", Amount: 110000, Type: TrxType(reconcile.Debit), TransactionTime: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)},
						Bank:        BankStatement{UniqueID: "BCA-1", Amount: -110000, Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Bank: "BCA"},
						Discrepancy: 0,
					},
				},
				UnmatchedSystem: []SystemTransaction{},
				UnmatchedBank:   []UnmatchedByBank{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			reconcileUsecase := NewMockreconcileUsecaseManager(ctrl)

			mockFields := mockFields{
				reconcile: reconcileUsecase,
			}
			test.mock(mockFields)

			h := &Handler{
				reconcile: mockFields.reconcile,
			}

			req := httptest.NewRequest(test.args.method, "/reconcile", strings.NewReader(test.args.body))
			rec := httptest.NewRecorder()

			h.HandleReconcile(rec, req)

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
