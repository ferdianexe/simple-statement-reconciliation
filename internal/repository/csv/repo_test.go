package csv

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_NewRepository(t *testing.T) {
	type args struct {
		infra infraProvider
	}
	tests := []struct {
		name string
		args func(ctrl *gomock.Controller) args
		want func(a args) *Repository
	}{
		{
			name: "success",
			args: func(ctrl *gomock.Controller) args {
				return args{infra: NewMockinfraProvider(ctrl)}
			},
			want: func(a args) *Repository {
				return &Repository{infra: a.infra}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			a := test.args(ctrl)

			got := NewRepository(a.infra)

			assert.NotNil(t, got)
			assert.Equal(t, test.want(a), got)
		})
	}
}
