package transaction

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewResource(t *testing.T) {
	type args struct {
		csvRepo csvRepoProvider
	}
	tests := []struct {
		name string
		args func(ctrl *gomock.Controller) args
		want func(a args) *Resource
	}{
		{
			name: "success",
			args: func(ctrl *gomock.Controller) args {
				return args{
					csvRepo: NewMockcsvRepoProvider(ctrl),
				}
			},
			want: func(a args) *Resource {
				return &Resource{csvRepo: a.csvRepo}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			a := test.args(ctrl)

			got := NewResource(a.csvRepo)

			assert.NotNil(t, got)
			assert.Equal(t, test.want(a), got)
		})
	}
}
