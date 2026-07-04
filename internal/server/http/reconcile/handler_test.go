package reconcile

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		reconcile reconcileUsecaseManager
		infra     infraProvider
	}
	tests := []struct {
		name string
		args func(ctrl *gomock.Controller) args
		want func(a args) *Handler
	}{
		{
			name: "success",
			args: func(ctrl *gomock.Controller) args {
				return args{
					reconcile: NewMockreconcileUsecaseManager(ctrl),
					infra:     NewMockinfraProvider(ctrl),
				}
			},
			want: func(a args) *Handler {
				return &Handler{reconcile: a.reconcile, infra: a.infra}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			a := test.args(ctrl)

			got := NewHandler(a.reconcile, a.infra)

			assert.NotNil(t, got)
			assert.Equal(t, test.want(a), got)
		})
	}
}
