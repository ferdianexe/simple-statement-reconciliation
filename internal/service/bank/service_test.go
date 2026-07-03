package bank

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	type args struct {
		resource resourceProvider
	}
	tests := []struct {
		name string
		args func(ctrl *gomock.Controller) args
		want func(a args) *Service
	}{
		{
			name: "success",
			args: func(ctrl *gomock.Controller) args {
				return args{
					resource: NewMockresourceProvider(ctrl),
				}
			},
			want: func(a args) *Service {
				return &Service{resource: a.resource}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			a := test.args(ctrl)

			got := NewService(a.resource)

			assert.NotNil(t, got)
			assert.Equal(t, test.want(a), got)
		})
	}
}
