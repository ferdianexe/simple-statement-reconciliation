package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	type args struct {
		handlers Handlers
	}
	tests := []struct {
		name string
		args func(ctrl *gomock.Controller) args
		want func(a args) *Server
	}{
		{
			name: "success",
			args: func(ctrl *gomock.Controller) args {
				return args{
					handlers: Handlers{Reconcile: NewMockreconcileHandler(ctrl)},
				}
			},
			want: func(a args) *Server {
				return &Server{handlers: a.handlers}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			a := test.args(ctrl)

			got := NewServer(a.handlers)

			assert.NotNil(t, got)
			assert.Equal(t, test.want(a), got)
		})
	}
}

func TestServer_NewMux(t *testing.T) {
	type mockFields struct {
		reconcile *MockreconcileHandler
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name       string
		mock       func(m mockFields)
		args       args
		wantStatus int
		wantBody   string
	}{
		{
			name: "GET_healthz_returns_200_ok_without_touching_the_reconcile_handler",
			mock: func(m mockFields) {},
			args: args{
				method: http.MethodGet,
				path:   "/healthz",
			},
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name: "POST_reconcile_delegates_to_the_reconcile_handler",
			mock: func(m mockFields) {
				m.reconcile.EXPECT().
					HandleReconcile(gomock.Any(), gomock.Any()).
					Times(1)
			},
			args: args{
				method: http.MethodPost,
				path:   "/reconcile",
			},
			wantStatus: http.StatusOK, // mock does nothing, so the default 200 stands
			wantBody:   "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			reconcileHandler := NewMockreconcileHandler(ctrl)

			mockFields := mockFields{
				reconcile: reconcileHandler,
			}
			test.mock(mockFields)

			srv := NewServer(Handlers{Reconcile: mockFields.reconcile})

			req := httptest.NewRequest(test.args.method, test.args.path, nil)
			rec := httptest.NewRecorder()

			srv.NewMux().ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Code)
			if test.wantBody != "" {
				assert.Equal(t, test.wantBody, rec.Body.String())
			}
		})
	}
}

func TestHandleHealthz(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "always_returns_200_ok",
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()

			handleHealthz(rec, req)

			assert.Equal(t, test.wantStatus, rec.Code)
			assert.Equal(t, test.wantBody, rec.Body.String())
		})
	}
}

// Run() itself (the ListenAndServe wrapper) is intentionally not
// unit-tested here: it's a two-line pass-through to the standard
// library's blocking http.ListenAndServe, and exercising it would mean
// binding a real port and racing a shutdown rather than testing any
// logic that belongs to this package.
